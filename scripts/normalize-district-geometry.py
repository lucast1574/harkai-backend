"""Repair reference polygons offline. Runtime remains Go-only with embedded geometry."""
import argparse
import gzip
import hashlib
import json
import pathlib
from shapely import make_valid
from shapely.geometry import shape, mapping
from shapely.geometry.polygon import orient


def polygons(geometry):
    if geometry.geom_type == 'Polygon':
        return [geometry]
    return [polygon for member in geometry.geoms for polygon in polygons(member)] if hasattr(geometry, 'geoms') else []


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('dataset', type=pathlib.Path)
    args = parser.parse_args()
    original = args.dataset.read_bytes()
    collection = json.loads(gzip.decompress(original))
    result, repaired = [], []
    for feature in collection['features']:
        identifier = feature['properties']['IDDIST']
        geometry = shape(feature['geometry'])
        if identifier[:4] not in {'1501', '0701', '1301'} or geometry.is_valid:
            result.append(feature)
            continue
        parts = polygons(make_valid(geometry))
        if not parts:
            raise ValueError('No district polygon after repair: ' + identifier)
        repaired.append({'district_id': identifier, 'parts': len(parts)})
        for part in parts:
            result.append({**feature, 'geometry': mapping(orient(part, sign=1))})
    if not repaired:
        print('No invalid polygons in the supported cities; dataset unchanged')
        return 0
    collection['features'] = result
    output = json.dumps(collection, ensure_ascii=False, separators=(',', ':')).encode()
    args.dataset.write_bytes(gzip.compress(output, mtime=0))
    manifest = {'input_sha256': hashlib.sha256(original).hexdigest(),
                'output_sha256': hashlib.sha256(args.dataset.read_bytes()).hexdigest(),
                'method': 'GEOS make_valid, retaining polygon parts and holes; WGS84 unchanged',
                'repaired': repaired}
    args.dataset.with_name('geometry-repairs.json').write_text(json.dumps(manifest, indent=2) + '\n')
    print(json.dumps(manifest))
    return len(repaired)


if __name__ == '__main__':
    main()
