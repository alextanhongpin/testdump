-- query --
SELECT * FROM users WHERE name = $1 AND age = $2

-- args --
{
 "$1": "John",
 "$2": 13
}

