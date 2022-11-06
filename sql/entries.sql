-- name: entries/get-all
SELECT
	id,
	timestamp,
	sold_quantity
FROM
	entries;

-- name: entries/create
INSERT INTO
	entries(sold_quantity)
VALUES
	(?1);
