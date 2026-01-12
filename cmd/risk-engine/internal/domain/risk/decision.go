func Decide(score int) string {
    if score >= 60 {
        return "BLOCKED"
    }
    return "APPROVED"
}
