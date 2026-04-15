// log_plugin.pb
// This plugin logs events but only if they contain certain keywords

func on_event(event) {
    if len(event["message"]) > 10 {
        notify_host("Caught a long message!")
        print("Plugin [Log]: Long message received:", event["message"])
        print("I dont know how it works but it works")
        return true
    }
    return false
}

// Global variable that the host can read
plugin_name := "Logger Pro"
