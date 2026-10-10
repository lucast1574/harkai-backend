# District reference cartography

Source: [SENACE ArcGIS layer 98](https://geosenace.senace.gob.pe/arcgis/rest/services/EVA_SERVER/Limite_distrital_2025/MapServer/98).
The service name contains 2025, but its published description identifies INEI district surfaces as of **June 2020**. This is a reference map, not current legal boundary certification. Recently created districts (including Alto Trujillo) need updated official geometry before being added; do not invent boundaries.

Downloaded with `where=1=1`, fields `IDDIST,NOMBDIST,NOMBPROV,NOMBDEP`, `outSR=4326`, `f=geojson`, `maxAllowableOffset=0.00005`, `geometryPrecision=7`, `returnGeometry=true`; gzip compressed. Simplification is approximately five metres. There are 1886 features; separate islands/territorial fragments can share an ID. Province choices deduplicate IDs; area queries aggregate all fragments.

The gzip is embedded in the Go binary and loaded before accepting requests. Reverse lookup uses bounding boxes then point-in-polygon with hole and edge handling. MongoDB uses its existing 2dsphere index for actual geographic filtering; client strings are not trusted as boundaries. The web caches up to 64 successful zone resolutions and uses session storage for its chosen consultation zone; the actual browser location stays in memory and is requested only by user action.

SHA256: `e27170d36ebad1ff0c0499c8c6ff32c40386764a83b109e131d56c306151a52c`.

Update the source and reference year together. Validate district/city fixtures and MongoDB polygon queries after replacing geometry.
