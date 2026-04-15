// filter_logs simulates a log processing plugin
func filter_logs(msg, level) {
    if level == "DEBUG" {
        return "" // suppress debug logs
    }
    
    if level == "ERROR" {
        return "!!! ALERT: " + msg
    }
    
    return msg
}

print(filter_logs("System starting", "INFO"))
print(filter_logs("Low disk space", "ERROR"))
print(filter_logs("Variable x is 5", "DEBUG"))
