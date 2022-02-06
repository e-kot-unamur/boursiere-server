-- name: users/get-all
SELECT
	id,
	name,
	password,
	admin
FROM
	users

-- name: users/count
SELECT
	COUNT(*)
FROM
	users

-- name: users/get-by-id
SELECT
	id,
	name,
	password,
	admin
FROM
	users
WHERE
	id = ?1

-- name: users/get-by-name
SELECT
	id,
	name,
	password,
	admin
FROM
	users
WHERE
	name = ?1

-- name: users/get-by-token
SELECT
	u.id,
	u.name,
	u.password,
	u.admin
FROM
	tokens AS t
INNER JOIN
	users AS u ON u.id = t.user_id
WHERE
	t.value = ?1

-- name: users/create
INSERT INTO
	users(name, password, admin)
VALUES
	(?1, ?2, ?3)

-- name: users/update
UPDATE
	users
SET
	name = ?2,
	password = ?3,
	admin = ?4
WHERE
	id = ?1

-- name: users/delete
DELETE FROM
	users
WHERE
	id = ?1

-- name: users/create-token
INSERT INTO
	tokens(value, user_id)
VALUES
	(?2, ?1)

-- name: users/delete-token
DELETE FROM
	tokens
WHERE
	value = ?1
