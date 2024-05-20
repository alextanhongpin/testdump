-- query --
SELECT *
  FROM users
  WHERE `name` = :v1
    AND age = :v2

-- args --
{
 "v1": "John",
 "v2": 13
}

