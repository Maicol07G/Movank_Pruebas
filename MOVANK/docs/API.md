# API

## Health
`GET /health`

## Products
`GET /v1/products`
`POST /v1/products`

Body:
```json
{"name":"Café","price_cents":5000}
```

## Sales
`POST /v1/sales`

Header:
`Idempotency-Key: <unique-operation-key>`

Body:
```json
{"product_id":"UUID","quantity":1}
```

A repeated request with the same key returns the existing sale instead of creating another one.

## Dashboard
`GET /v1/dashboard/today`

## Payment contract
Planned endpoint:
`POST /v1/sales/:id/pay`

Body:
```json
{"method":"CARD","scenario":"APPROVED"}
```

Scenarios:
- APPROVED
- DECLINED
- TIMEOUT -> UNKNOWN

## SSE
Planned endpoint:
`GET /v1/dashboard/stream`
