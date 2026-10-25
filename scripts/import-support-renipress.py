"""Build the permanent directory from the official RENIPRESS CSV (no live dependency)."""
import argparse
import collections
import csv
import datetime
import hashlib
import json
import pathlib
import re
import math
from urllib.parse import urlparse, parse_qs

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
CURATED = {'00006207': 'lima-loayza', '00006206': 'lima-dosdemayo',
           '00006197': 'lima-san-isidro', '00005196': 'trujillo-regional',
           '00005195': 'trujillo-belen', '00005204': 'trujillo-noria',
           '00005201': 'trujillo-aranjuez', '00005199': 'trujillo-union'}


def phone(raw, city):
    # Never concatenate multiple numbers, extensions or ambiguous fragments.
    match = re.fullmatch(r'(?:TEL\s*:\s*)?(\+?[\d ()-]+)', raw.strip(), re.I)
    if not match:
        return ''
    digits = re.sub(r'\D', '', match[1])
    if len(digits) == 9 and digits.startswith('9'):
        return '+51' + digits
    if (len(digits) == 9 and digits.startswith('01')) or (len(digits) == 9 and digits.startswith('0') and not digits.startswith('00')):
        return '+51' + digits[1:]
    if len(digits) == 7 and city == 'lima':
        return '+511' + digits
    if len(digits) == 6 and city == 'trujillo':
        return '+5144' + digits
    if (len(digits) == 11 and digits.startswith('519')) or (len(digits) == 10 and digits.startswith('51') and digits[2] != '0'):
        return '+' + digits
    return ''


def phones(raw, city):
    return list(dict.fromkeys(number for fragment in re.split(r'[/;,\n]', raw)
                             if (number := phone(fragment, city))))


def photos(row):
    result = []
    for key in ('IMAGEN_1', 'IMAGEN_2', 'IMAGEN_3'):
        raw = row.get(key, '').strip()
        parsed = urlparse(raw)
        query = parse_qs(parsed.query)
        if (parsed.scheme == 'http' and parsed.netloc == 'app20.susalud.gob.pe:8080'
                and parsed.path == '/registro-renipress-webapp/ipress.htm'
                and query.get('action') == ['downloadFile']
                and re.fullmatch(r'[\w.-]+\.(?:jpe?g|png|webp)', query.get('idArchivo', [''])[0], re.I)):
            result.append(raw)
    return list(dict.fromkeys(result))


def place(row, reviewed):
    province_id = row['UBIGEO'].zfill(6)[:4]
    city = {'1501': 'lima', '0701': 'lima', '1301': 'trujillo'}.get(province_id, province_id)
    kind = CLASSES.get(row['CLASIFICACION'])
    if not re.fullmatch(r'\d{6}', row['UBIGEO'].zfill(6)) or not kind or row['ESTADO'] != 'ACTIVO':
        return None
    code = row['COD_IPRESS'].zfill(8)
    if code in CURATED:
        return None
    try:
        latitude, longitude = float(row['NORTE']), float(row['ESTE'])
    except ValueError:
        return None
    bounds = (-18.5, 0, -82, -68)
    if not (math.isfinite(latitude) and math.isfinite(longitude) and bounds[0] < latitude < bounds[1] and bounds[2] < longitude < bounds[3]):
        return None
    address = row['DIRECCION'].strip()
    if not address or address.upper() in {'ACTUALIZAR', 'NO REGISTRA', 'SIN DATOS'}:
        return None
    contacts = phones(row['TELEFONO'], city)
    return {
        'id': 'renipress-' + code, 'registry_id': code,
        'name': row['NOMBRE'].strip(), 'kind': kind, 'city': city,
        'department': row['DEPARTAMENTO'].strip().title(), 'province': row['PROVINCIA'].strip().title(),
        'district': row['DISTRITO'].strip().title(), 'district_id': row['UBIGEO'].zfill(6),
        'address': address, 'latitude': latitude, 'longitude': longitude,
        'phone': contacts[0] if contacts else '', 'phones': contacts,
        'photo_urls': photos(row), 'institution': row['INSTITUCION'],
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
    curated_media = {}
    for row in rows:
        code = row['COD_IPRESS'].zfill(8)
        if code in CURATED:
            curated_media[CURATED[code]] = (code, photos(row), row['UBIGEO'].zfill(6))
        result = place(row, args.reviewed)
        if result:
            if result['id'] in places:
                raise ValueError('duplicate registry ID: ' + result['id'])
            places[result['id']] = result
    items = sorted(places.values(), key=lambda p: (p['city'], p['district'], p['name']))
    target = ROOT / 'internal/mongostore/support-seed'
    curated = json.loads((target / 'places.json').read_text(encoding='utf-8'))
    for item in curated:
        if item['id'] in curated_media:
            code, images, district_id = curated_media[item['id']]
            item.setdefault('registry_id', code)
            item.setdefault('district_id', district_id)
            if not item.get('photo_urls') and images:
                item['photo_urls'] = images
    (target / 'places.json').write_text(json.dumps(curated, ensure_ascii=False, indent=2) + '\n', encoding='utf-8')
    (target / 'renipress.json').write_text(json.dumps(items, ensure_ascii=False, indent=2) + '\n', encoding='utf-8')
    summary = {'source_url': SOURCE, 'source_updated_at': '2026-09-30',
               'reviewed_at': args.reviewed, 'input_sha256': hashlib.sha256(raw).hexdigest(),
               'imported': len(items), 'by_city': dict(collections.Counter(p['city'] for p in items)),
               'by_department': dict(sorted(collections.Counter(p['department'] for p in items).items())),
               'provinces': len({p['district_id'][:4] for p in items}),
               'districts': dict(collections.Counter(p['city'] for p in {p['district_id']: p for p in items}.values()))}
    (target / 'renipress-manifest.json').write_text(json.dumps(summary, indent=2) + '\n', encoding='utf-8')
    print(json.dumps(summary, ensure_ascii=False))
    return len(items)


if __name__ == '__main__':
    main()
