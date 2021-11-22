package messageboards.DELETE.boards.__boardID.messages.__messageID

import input.user.applications as applications

default allowed = false


allowed {
    input.user.attributes.roles[i] == "operations"
}

allowed {
    input.user.identities[input.resource.board.owner]
}

allowed {
    input.user.identities[input.resource.message.sender]
}
