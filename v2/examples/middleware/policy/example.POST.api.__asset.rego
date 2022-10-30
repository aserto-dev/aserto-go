package example.POST.api.__asset

import future.keywords.in

default allowed = false

allowed {
    input.user.attributes.roles[_] == "editor"
    input.resource.asset != "secret"
}

allowed {
    input.user.attributes.roles[_] == "admin"
}
