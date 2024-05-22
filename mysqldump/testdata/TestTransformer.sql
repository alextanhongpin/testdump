-- query --
SELECT *
  FROM users
  WHERE `name` = :v1
    AND id = :v2

-- args --
{
 ":v1": "John",
 ":v2": 1
}

