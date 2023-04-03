SELECT
  delivery_id,
  signature,
  received,
  event,
  payload
FROM
  `${dataset_id}.${table_id}`
GROUP BY
  delivery_id,
  signature,
  received,
  event,
  payload;
