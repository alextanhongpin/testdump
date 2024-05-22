-- query --
select * from users where `name` = :v1 and created_at > :v2

-- args --
{
 ":v1": "John",
 ":v2": "2024-05-22T21:34:26.810862+08:00"
}

