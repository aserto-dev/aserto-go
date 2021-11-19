package messageboards.POST.boards

default allowed = false

allowed {
    input.user  # must be authenticated to create a message board
}
