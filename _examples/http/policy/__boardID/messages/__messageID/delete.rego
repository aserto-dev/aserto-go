package messageboards.DELETE.boards.__boardID.messages.__messageID

import input.user.applications as applications

default allowed = false


allowed {
    input.user.attributes.roles[i] == "operations"
}
