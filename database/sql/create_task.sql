INSERT INTO task (sign, name, expected, finished, model, txHash, startDate, finishDate, epochID)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, (SELECT MAX(id) FROM epoch));