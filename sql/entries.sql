-- name: entries/get-all
SELECT
	id,
	timestamp,
	sold_quantity,
	endOfParty
FROM
	entries;

-- name: entries/create
INSERT INTO
	entries(sold_quantity, endOfParty)
VALUES
	(?1, ?2);

-- name: entries/delete-all
DELETE FROM
	entries

-- name: entries/stat/currentPeople
SELECT SUM(sold_quantity) FROM entries WHERE endOfParty != true

-- name: entries/stat/allEntries
SELECT SUM(sold_quantity) FROM entries
