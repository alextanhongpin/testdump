package test
// https://cuelang.org/docs/howto/

import "list"
import "time"
import "strings"

let url = =~ "^https://(.+)"

#User: close({
    name!: string & strings.MinRunes(2) & strings.MaxRunes(8)
    age!: >= 13
    hobbies!: [...string] & list.MinItems(1) & list.MaxItems(1)
    birthday!: string & time.Format("2006-01-02")
    imageURL: string & url
})
