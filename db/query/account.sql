-- name: CreateAccount :one
insert into accounts(
owner,
balance,
currency

)values($1,$2,$3) returning *;

-- name: GetAccount :one
select * from accounts
where "id" =$1 limit 1;

-- name: GetAccountForUpdate :one
select * from accounts
where "id" =$1 limit 1
FOR NO KEY UPDATE;



-- name: ListAccounts :many
SELECT * FROM accounts
WHERE owner = $1
ORDER BY id
LIMIT $2
OFFSET $3;

-- name: UpadateAccount :one
update accounts set balance=$2
where "id"=$1 returning *;

-- name: AddAccountBalance :one
UPDATE accounts
SET balance=balance + sqlc.arg(amount)
WHERE id = sqlc.arg(id)
RETURNING *;


-- name: DeleteAccounts :exec
delete from accounts
where "id"=$1;