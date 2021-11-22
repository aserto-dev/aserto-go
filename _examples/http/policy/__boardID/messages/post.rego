package messageboards.POST.boards.__boardID.messages

default allowed = false

allowed {
    input.user  # must be authenticated to post messages
}
