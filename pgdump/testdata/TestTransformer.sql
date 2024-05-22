-- query --
SELECT *
  FROM users
  WHERE name = $1
    AND id = $2

-- args --
{
 "$1": "John",
 "$2": 1
}

