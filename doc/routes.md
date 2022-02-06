# Routes

## Authentication

Authentication is made using an access token. It is a random base64 string that identifies its owner. Use the `POST /api/users/token` with your username/password to create one. `DELETE` on the same path can be used to remove your token from the database.

To use a token in an authenticated route, use the HTTP *Authorization* header with a *Bearer* type:

```http
POST /api/beers/order HTTP/1.1
…
Authorization: Bearer Xepk1c6fhGr5ItJeZeM6PmJjz2s…
…
```

## Common responses

All routes have responses for invalid, unauthenticated or unauthorized requests. There are described here once and for all.

400 Bad Request

```json
{
  "error": "bad_request"
}
```

401 Unauthorized

```json
{
  "error": "unauthenticated"
}
```

403 Forbidden

```json
{
  "error": "unauthorized"
}
```

## GET /api/beers

Get the current status of all beers (ID, bar, name, quantity, price, etc.).

### Responses

200 OK

```json
[
  {
    "id": 1,
    "barId": 1,
    "name": "Bush (33cL)",
    "stockQuantity": 24,
    "totalSoldQuantity": 4,
    "sellingPrice": 1.2,
    "previousSellingPrice": 1.1,
    "bottleSize": 33,
    "alcoholContent": 12
  },
  …
]
```

## POST /api/beers

Upload beer types to the database. An admin access token is required.

Existing beers and their history are deleted beforehand.

### Request

`text/csv`

```csv
name,barId,bottleSize,alcoholContent,Prix/Bac,Bouteilles/Bac,Nb. Bacs,stockQuantity,purchasePrice,incrCoef,decrCoef,minCoef,maxCoef,Prix minimum,Prix maximum
Barbar,2,33,8,"23,66 €",20,3,60,"2,54 €","0,07 €","0,09 €","0,85","2,5","2,20 €","6,30 €"
Bertinchamps Triple,2,50,8,"49,39 €",20,2,40,"0,81 €","0,08 €","0,08 €","0,85","2,5","0,70 €","2,00 €"
```

### Responses

201 Created

```json
[
  {
    "id": 1,
    "barId": 2,
    "name": "Barbar",
    "stockQuantity": 60,
    "totalSoldQuantity": 0,
    "sellingPrice": 2.54,
    "previousSellingPrice": 2.54,
    "bottleSize": 33,
    "alcoholContent": 8
  },
  {
    "id": 2,
    "barId": 2,
    "name": "Bertinchamps Triple",
    "stockQuantity": 40,
    "totalSoldQuantity": 0,
    "sellingPrice": 0.81,
    "previousSellingPrice": 0.81,
    "bottleSize": 50,
    "alcoholContent": 8
  }
]
```

## GET /api/beers/events

TODO

## POST /api/beers/order

Add (or remove) an amount to beers' sold quantities. An access token is required.

Please note that invalid IDs are simply ignored.

### Request

```json
[
  {
    "id": 1,
    "orderedQuantity": 2
  },
  {
    "id": 4,
    "orderedQuantity": -1
  },
  …
]
```

### Responses

204 No Content

## GET /api/beers/stats

Get statistics about the event that are shown on the administrator page. An admin access token is required.

### Responses

200 OK

```json
{
  "estimatedProfit": 1836.4
}
```

## GET /api/users

Return a list of every user. An admin access token is required.

### Responses

200 OK

```json
[
  {
    "id": 1,
    "name": "admin",
    "admin": true
  },
  …
]
```

## POST /api/users

Create a new user. An admin access token is required.

### Request

```json
{
  "name": "marcel",
  "password": "asuperpassword",
  "admin": false
}
```

### Responses

201 Created

```json
{
  "id": 2,
  "name": "marcel",
  "admin": false
}
```

400 Bad Request

```json
{
  "error": "non_unique_name"
}
```

## PATCH /api/users/:id

Edit user information. An access token is required. You must indeed be authenticated as an administrator or as the concerned user.


This is a `PATCH` route: only provided fields are updated, the others are left as is.

In addition, only an administrator can give or revoke admin permissions.

### Request

```json
{
  "admin": true
}
```

### Responses

200 OK

```json
{
  "id": 2,
  "name": "marcel",
  "admin": true
}
```

400 Bad Request

```json
{
  "error": "non_unique_name"
}
```

404 Not Found

```json
{
  "error": "invalid_id"
}
```

## DELETE /api/users/:id

Delete user with a specific ID. An admin access token is required.

### Responses

204 No Content

404 Not Found

```json
{
  "error": "invalid_id"
}
```

## POST /api/users/token

Generate a new access token from username and password.

### Request

```json
{
  "name": "admin",
  "password": "passwordyword"
}
```

### Responses

201 Created

```json
{
  "id": 1,
  "name": "admin",
  "admin": true,
  "token": "Xepk1c6fhGr5ItJeZeM6PmJjz2s…"
}
```

401 Unauthorized

```json
{
  "error": "wrong_credentials"
}
```

## DELETE /api/users/token

Delete a given access token, effectively logging out. The deleted token is the one contained in the *Authorization* header

### Responses

204 No Content
