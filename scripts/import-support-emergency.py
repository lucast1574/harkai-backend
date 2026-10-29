"""Snapshot official police/fire station coordinates; never depend on live GIS at runtime."""
import collections
import datetime
import gzip
import hashlib
import json
import math
import pathlib
import unicodedata
import urllib.parse
import urllib.request

ROOT = pathlib.Path(__file__).resolve().parents[1]
SOURCES = {
    'police': 'https://geomininter.mininter.gob.pe/arcgis/rest/services/pnp/policia_nacional_peru/MapServer/0',
    'fire_station': 'https://www.idep.gob.pe/geoportal/rest/services/INSTITUCIONALES/INFRAESTRUCTURA_SOCIAL/FeatureServer/3',
}


def normalized(value):
    value = ''.join(c for c in unicodedata.normalize('NFD', value.upper().strip())
                   if unicodedata.category(c) != 'Mn')
    return {'NAZCA': 'NASCA', 'LA TINGUIAA': 'LA TINGUINA', 'BREAA': 'BRENA',
            'FERREYAFE': 'FERRENAFE', 'LOS BAROS DEL INCA': 'LOS BANOS DEL INCA',
            'CAMETE': 'CANETE', 'SAN VICENTE DE CADETE': 'SAN VICENTE DE CANETE',
            'PARIRAS': 'PARINAS'}.get(value, value)


def snapshot(source):
    features = []
    offset = 0
    while True:
        query = urllib.parse.urlencode(dict(where='1=1', outFields='*', outSR=4326,
                                           returnGeometry='true', resultOffset=offset,
                                           resultRecordCount=500, orderByFields='objectid', f='json'))
        with urllib.request.urlopen(source + '/query?' + query, timeout=90) as response:
            result = json.load(response)
        if 'error' in result:
            raise ValueError(result['error'])
        batch = result['features']
        features.extend(batch)
        if not result.get('exceededTransferLimit'):
            return features
        if not batch:
            raise ValueError('GIS pagination returned no progress')
        offset += len(batch)


def main():
    reviewed = datetime.date.today().isoformat()
    with gzip.open(ROOT / 'internal/districtgeo/data/peru-districts.geojson.gz') as handle:
        districts = [f['properties'] for f in json.load(handle)['features']]
    lookup = {(normalized(d['NOMBDEP']), normalized(d['NOMBPROV']), normalized(d['NOMBDIST'])): d for d in districts}
    by_id = {d['IDDIST']: d for d in districts}
    # Current ubigeos are also present in RENIPRESS, including Alto Trujillo.
    registry = json.loads((ROOT / 'internal/mongostore/support-seed/renipress.json').read_text(encoding='utf-8'))
    for p in registry:
        by_id.setdefault(p['district_id'], dict(IDDIST=p['district_id'], NOMBDEP=p['department'], NOMBPROV=p['province'], NOMBDIST=p['district']))
    places, manifests = [], {}
    for kind, source in SOURCES.items():
        features = snapshot(source)
        skipped = collections.Counter()
        for feature in features:
            a = {k.lower(): str(v or '').strip() for k, v in feature['attributes'].items()}
            geo = feature.get('geometry') or {}
            lat, lng = geo.get('y'), geo.get('x')
            if not all(isinstance(v, (float, int)) and math.isfinite(v) for v in (lat, lng)) or not (-19 < lat < 0 and -82 < lng < -68):
                skipped['coordinates'] += 1
                continue
            dep = a.get('departamento') or a.get('departamen', '')
            district = lookup.get(tuple(map(normalized, (dep, a['provincia'], a['distrito']))))
            if kind == 'police':
                district = by_id.get(a.get('id_dist')) or district
            if kind == 'fire_station' and (dep, a['provincia'], a['distrito']) == ('PIURA', 'LA BREA', 'NEGRITOS'):
                district = by_id['200705']  # La Brea district, Talara province; Negritos is its capital.
            if not district:
                skipped['unmatched_district'] += 1
                continue
            name = a.get('comisaria') or a.get('nombre')
            if not name:
                skipped['missing_name'] += 1
                continue
            ubigeo = district['IDDIST']
            province = ubigeo[:4]
            places.append(dict(
                id=f"{kind}-{a['objectid']}", name=name, kind=kind,
                city={'1501': 'lima', '0701': 'lima', '1301': 'trujillo'}.get(province, province),
                district_id=ubigeo, department=district['NOMBDEP'].title(),
                province=district['NOMBPROV'].title(), district=district['NOMBDIST'].title(),
                address='Dirección no publicada en la fuente; ubicación de referencia en el mapa',
                latitude=lat, longitude=lng,
                institution='Policía Nacional del Perú' if kind == 'police' else 'Cuerpo General de Bomberos Voluntarios del Perú',
                source_url=source, location_source_url=source, reviewed_at=reviewed,
            ))
        manifests[kind] = dict(source_url=source, retrieved=len(features), skipped=dict(skipped),
                               snapshot_sha256=hashlib.sha256(json.dumps(features, sort_keys=True).encode()).hexdigest())
    places.sort(key=lambda p: p['id'])
    target = ROOT / 'internal/mongostore/support-seed'
    (target / 'emergency.json').write_text(json.dumps(places, ensure_ascii=False, indent=2) + '\n', encoding='utf-8')
    manifest = dict(reviewed_at=reviewed, sources=manifests, imported=dict(collections.Counter(p['kind'] for p in places)))
    (target / 'emergency-manifest.json').write_text(json.dumps(manifest, ensure_ascii=False, indent=2) + '\n', encoding='utf-8')
    print(json.dumps(manifest, ensure_ascii=False))


if __name__ == '__main__':
    main()
