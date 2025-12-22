# FHIR Renderer API

Renders FHIR ResourceDefinition structures as SVG diagrams.

## Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | /health | Health check → {"status":"ok"} |
| GET | /help | This documentation |
| GET | /example | Example ResourceDefinition JSON |
| GET | /render?resource={compressed} | Render Brotli+Base64URL compressed JSON to SVG |
| POST | /render | Render JSON body to SVG |
| POST | /compress | Compress JSON → {"compressed": "..."} |
| POST | /decompress | Decompress {"data": "..."} → JSON |

## JSON Schema

### ResourceDefinition (root)
```json
{
  "resourceType": "ResourceDefinition",  // optional, identifier
  "name": "...",                          // REQUIRED: resource name
  "type": "...",                          // REQUIRED: e.g. "DomainResource"
  "flags": ["TU", "N"],                   // optional: metadata flags
  "description": "...",                   // optional: human description
  "elements": [...],                      // optional: child elements
  "extensions": [...]                     // optional: FHIR extensions
}
```

### Element (nested)
```json
{
  "name": "...",             // REQUIRED: field name
  "type": "...",             // REQUIRED: data type
  "cardinality": "0..*",     // optional: "0..1", "1..1", "0..*", "1..*"
  "flags": ["S", "?!"],      // optional: FHIR flags
  "typeRef": "https://...",  // optional: link to type docs
  "description": "...",      // optional: field description
  "usage": "used",           // optional: implementation status
  "notes": "...",            // optional: custom notes
  "binding": {...},          // optional: value set binding
  "elements": [...],         // optional: nested children (BackboneElement)
  "extensions": [...]        // optional: extensions on this element
}
```

### Binding
```json
{
  "strength": "required",              // "required"|"extensible"|"preferred"|"example"
  "valueSet": "val1 | val2 | val3",    // allowed values or reference URL
  "url": "https://..."                 // optional: value set docs link
}
```

### Extension
```json
{
  "name": "...",           // REQUIRED: extension name
  "url": "https://...",    // REQUIRED: extension URL
  "type": "...",           // REQUIRED: data type
  "cardinality": "0..1",   // optional: cardinality
  "context": "...",        // optional: where extension applies (root-level only)
  "description": "..."     // optional: description
}
```

## Flags

| Flag | Symbol | Meaning |
|------|--------|---------|
| S | Σ | Summary element |
| ?! | ?!Σ | Modifier element |
| I | I | Has constraint |
| TU | [TU] | Trial use (boxed) |
| N | [N] | Normative (boxed) |

## Usage Values

| Value | Rendering |
|-------|-----------|
| used | Normal style |
| not-used | Grayed out (#999) |
| todo | Bold orange, "TODO:" prefix |
| optional | Default style |

## Icons (auto-selected by type)

- **Folder (yellow)**: Root resource
- **Folder+dot**: BackboneElement (nested structure)
- **Diamond (blue)**: Simple element
- **Circle "E" (orange)**: Extension
- **Circle+line (green)**: Choice type [x]
- **Arrow (blue)**: Reference type

## Examples

### Compress JSON
```bash
curl -X POST http://localhost:8080/compress \
  -H "Content-Type: application/json" \
  -d '{"name":"Patient","type":"DomainResource"}'
# Returns: {"compressed":"G8gBgJOgxIm..."}
```

### GET Request (with compressed data)
```
GET /render?resource=G8gBgJOgxIm...
```

### POST Request (raw JSON)
```bash
curl -X POST http://localhost:8080/render \
  -H "Content-Type: application/json" \
  -d '{"name":"Patient","type":"DomainResource","elements":[{"name":"id","type":"id","cardinality":"0..1"}]}'
```

### Decompress
```bash
curl -X POST http://localhost:8080/decompress \
  -H "Content-Type: application/json" \
  -d '{"data":"G8gBgJOgxIm..."}'
# Returns: {"name":"Patient","type":"DomainResource"}
```

### Minimal Valid JSON
```json
{"name":"MyResource","type":"DomainResource"}
```

## Response

- **Success**: SVG/XML (Content-Type: image/svg+xml)
- **Error**: JSON with "error" and optional "details" fields

## Errors

| Code | Cause |
|------|-------|
| 400 | Missing 'resource' param, invalid JSON, missing 'name'/'type' |

## URL Compression

The GET /render endpoint uses Brotli compression + Base64URL encoding for ~60-70% size reduction.

**Format:** Brotli compress → Base64URL encode (no padding)

Use the /compress endpoint to create compressed strings, or use the interactive editor.

## Notes

- GET /render: Requires Brotli+Base64URL compressed JSON (use /compress or the editor)
- POST /render: Send raw JSON with Content-Type: application/json
- CORS enabled (Access-Control-Allow-Origin: *)
- Responses cached 1 hour (Cache-Control: public, max-age=3600)
