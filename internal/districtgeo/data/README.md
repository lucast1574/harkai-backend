# District reference cartography

Source: [SENACE ArcGIS layer 98](https://geosenace.senace.gob.pe/arcgis/rest/services/EVA_SERVER/Limite_distrital_2025/MapServer/98).
The service name contains 2025, but its published description identifies INEI district surfaces as of **June 2020**. This is a reference map, not current legal boundary certification. Recently created districts (including Alto Trujillo) need updated official geometry before being added; do not invent boundaries.

Downloaded with `where=1=1`, fields `IDDIST,NOMBDIST,NOMBPROV,NOMBDEP`, `outSR=4326`, `f=geojson`, `maxAllowableOffset=0.00005`, `geometryPrecision=7`, `returnGeometry=true`; gzip compressed. Simplification is approximately five metres. There were 1886 source features; the normalized dataset has 1888 polygon features; separate islands/territorial fragments can share an ID. Province choices deduplicate IDs; area queries aggregate all fragments.

The gzip is embedded in the Go binary and loaded before accepting requests. Reverse lookup uses bounding boxes then point-in-polygon with hole and edge handling. MongoDB uses its existing 2dsphere index for actual geographic filtering; client strings are not trusted as boundaries. The web caches up to 64 successful zone resolutions and uses session storage for its chosen consultation zone; the actual browser location stays in memory and is detected automatically once per page, with manual retry available.

Source gzip SHA256: `e27170d36ebad1ff0c0499c8c6ff32c40386764a83b109e131d56c306151a52c`.
Normalized gzip SHA256: `9a290a988fa20b1f8f5e6beb9d9066293b5b02fd30d9d1cc9a206d27cd7f6651`.

The Callao district (`070101`) had a self-intersecting reference ring that MongoDB rejected. Offline GEOS `make_valid` splits it into three valid polygon parts, retaining holes, geographic coordinates and district identity. See `geometry-repairs.json`. Reproduce from the source gzip with Python and Shapely 2.2.0: `python scripts/normalize-district-geometry.py internal/districtgeo/data/peru-districts.geojson.gz`. No Python or GEOS is needed at runtime. MongoDB integration verifies Lima/Callao city queries and both incident and support resource geofences. Lima and Callao form one metropolitan exploration area with 50 distinct district IDs; individual district queries remain separate.

Update the source and reference year together. Validate district/city fixtures and MongoDB polygon queries after replacing geometry.
