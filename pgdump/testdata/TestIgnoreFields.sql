-- query --
SELECT * FROM users WHERE name = $1 AND created_at > $2

-- args --
{
 "$1": "John",
 "$2": "2024-05-22T22:06:30.703188+08:00"
}

