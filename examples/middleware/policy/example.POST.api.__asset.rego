package example.POST.api.__asset

import future.keywords.in

default allowed = false

allowed {
    roles := {"editor", "admin"}
    some x in roles
    input.user.attributes.roles[_] == x
}
