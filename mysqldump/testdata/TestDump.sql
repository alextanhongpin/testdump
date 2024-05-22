-- query --
select * from users where `name` = :v1 and age = :v2

-- args --
{
 ":v1": "John",
 ":v2": 13
}

