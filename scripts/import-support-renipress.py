"""Build the permanent directory from the official RENIPRESS CSV (no live dependency)."""
import argparse
import collections
import csv
import datetime
import hashlib
import json
import pathlib
import re

ROOT = pathlib.Path(__file__).resolve().parents[1]
SOURCE = 'https://www.datosabiertos.gob.pe/dataset/registro-nacional-de-entidades-prestadoras-de-servicios-de-salud-renipress/resource/2d4389ed'
CLASSES = {
    'HOSPITALES O CLINICAS DE ATENCION ESPECIALIZADA': 'hospital',
    'HOSPITALES O CLINICAS DE ATENCION GENERAL': 'hospital',
    'INSTITUTOS DE SALUD ESPECIALIZADOS': 'hospital',
    'CENTROS DE SALUD O CENTROS MEDICOS': 'health_center',
    'CENTROS DE SALUD CON CAMAS DE INTERNAMIENTO': 'health_center',
    'PUESTOS DE SALUD O POSTAS DE SALUD': 'health_center',
    'POLICLINICOS': 'health_center',
    'CENTROS MEDICOS ESPECIALIZADOS': 'health_center',
}
# These already have curated records; preserve their coordinates and contacts.
CURATED = {'00006207', '00006206', '00006197', '00005196', '00005195', '00005204', '00005201', '00005199'}


def phone(raw, city):
    # Never concatenate multiple numbers, extensions or ambiguous fragments.
    match = re.fullmatch(r'(?:TEL\s*:\s*)?(\+?[\d ()-]+)', raw.strip(), re.I)
    if not match:
        return ''
    digits = re.sub(r'\D', '', match[1])
    if len(digits) == 9 and digits.startswith('9'):
        return '+51' + digits
    if (len(digits) == 9 and digits.startswith('01')) or (len(digits) == 9 and digits.startswith('044')):
        return '+51' + digits[1:]
    if len(digits) == 7 and city == 'lima':
        return '+511' + digits
    if len(digits) == 6 and city == 'trujillo':
        return '+5144' + digits
    if len(digits) == 11 and digits.startswith('51'):
        return '+' + digits
    return ''


def place(row, reviewed):
    city = {'1501': 'lima', '0701': 'lima', '1301': 'trujillo'}.get(row['UBIGEO'].zfill(6)[:4])
    kind = CLASSES.get(row['CLASIFICACION'])
    if not city or not kind or row['ESTADO'] != 'ACTIVO':
        return None
    code = row['COD_IPRESS'].zfill(8)
    if code in CURATED:
        return None
    try:
        latitude, longitude = float(row['NORTE']), float(row['ESTE'])
    except ValueError:
        return None
    bounds = (-13, -11, -78, -76) if city == 'lima' else (-9, -7, -80, -78)
    if not (bounds[0] < latitude < bounds[1] and bounds[2] < longitude < bounds[3]):
        return None
    address = row['DIRECCION'].strip()
    if not address or address.upper() in {'ACTUALIZAR', 'NO REGISTRA', 'SIN DATOS'}:
        return None
    return {
        'id': 'renipress-' + code, 'registry_id': code,
        'name': row['NOMBRE'].strip(), 'kind': kind, 'city': city,
        'district': row['DISTRITO'].strip().title(), 'district_id': row['UBIGEO'].zfill(6),
        'address': address, 'latitude': latitude, 'longitude': longitude,
        'phone': phone(row['TELEFONO'], city), 'institution': row['INSTITUCION'],
        'classification': row['CLASIFICACION'], 'category': row['CATEGORIA'],
        'source_url': SOURCE, 'location_source_url': SOURCE,
        'source_updated_at': '2026-09-30', 'reviewed_at': reviewed,
    }


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('csv', type=pathlib.Path)
    parser.add_argument('--reviewed', required=True)
    args = parser.parse_args()
    datetime.date.fromisoformat(args.reviewed)
    raw = args.csv.read_bytes()
    rows = csv.DictReader(raw.decode('utf-8-sig').splitlines(), delimiter=';')
    places = {}
    for row in rows:
        result = place(row, args.reviewed)
        if result:
            if result['id'] in places:
                raise ValueError('duplicate registry ID: ' + result['id'])
            places[result['id']] = result
    items = sorted(places.values(), key=lambda p: (p['city'], p['district'], p['name']))
    target = ROOT / 'internal/mongostore/support-seed'
    (target / 'renipress.json').write_text(json.dumps(items, ensure_ascii=False, indent=2) + '\n')
    summary = {'source_url': SOURCE, 'source_updated_at': '2026-09-30',
               'reviewed_at': args.reviewed, 'input_sha256': hashlib.sha256(raw).hexdigest(),
               'imported': len(items), 'by_city': dict(collections.Counter(p['city'] for p in items)),
               'districts': dict(collections.Counter(p['city'] for p in {p['district_id']: p for p in items}.values()))}
    (target / 'renipress-manifest.json').write_text(json.dumps(summary, indent=2) + '\n')
    print(json.dumps(summary, ensure_ascii=False))
    return len(items)


if __name__ == '__main__':
    main()
