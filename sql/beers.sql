-- name: beers/get-all
WITH
	h1 AS (
		SELECT
			beer_id,
			sold_quantity,
			selling_price,
			SUM(sold_quantity) AS total_sold_quantity,
			MAX(timestamp) AS most_recent_timestamp
		FROM
			history
		GROUP BY
			beer_id
	),
	h2 AS (
		SELECT
			h.beer_id,
			h.sold_quantity,
			h.selling_price,
			MAX(h.timestamp) AS second_most_recent_timestamp
		FROM
			history AS h
		INNER JOIN
			h1 USING (beer_id)
		WHERE
			h.timestamp < h1.most_recent_timestamp
		GROUP BY
			h.beer_id
	)
SELECT
	b.id,
	b.bar_id,
	b.name,
	b.stock_quantity,
	COALESCE(h1.sold_quantity, 0) AS sold_quantity,
	COALESCE(h2.sold_quantity, 0) AS previous_sold_quantity,
	COALESCE(h1.total_sold_quantity, 0) AS total_sold_quantity,
	COALESCE(h1.selling_price, b.purchase_price) AS selling_price,
	COALESCE(h2.selling_price, h1.selling_price, b.purchase_price) AS previous_selling_price,
	b.purchase_price,
	b.alcohol_content,
	b.incr_coef,
	b.decr_coef,
	b.min_coef,
	b.max_coef
FROM
	beers AS b
LEFT JOIN
	h1 ON b.id = h1.beer_id
LEFT JOIN
	h2 ON b.id = h2.beer_id

-- name: beers/get-estimated-profit
SELECT
	COALESCE(SUM(h.sold_quantity * (ROUND(h.selling_price, 1) - b.purchase_price)), 0) AS estimated_profit
FROM
	history AS h
INNER JOIN
	beers AS b ON b.id = h.beer_id

-- name: beers/make-order
UPDATE
	history
SET
	sold_quantity = sold_quantity + ?2
WHERE
	beer_id = ?1
	AND timestamp = (SELECT MAX(timestamp) FROM history WHERE beer_id = ?1)

-- name: beers/update-price
INSERT INTO
	history(beer_id, sold_quantity, selling_price)
VALUES
	(?1, 0, ?2)
