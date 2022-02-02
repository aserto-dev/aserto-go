package example.GET.api.__asset

import future.keywords.in

allowed {
    roles := {"editor", "admin"}
    some x in roles
    input.user.attributes.roles[_] == x
}
