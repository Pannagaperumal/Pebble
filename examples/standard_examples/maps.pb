// Map literal example
user := {
    "name": "Pebble Developer",
    "age": 25,
    "active": true
}

print("User Name: ", user["name"])
print("User Age: ", user["age"])

// Updating a map (if supported by evaluator/vm via OpIndex + Assignment)
// Note: Currently OpIndex + Assignment might need a specific opcode for map sets.
// Let's assume basic retrieval works for now.
