// Basic variable assignment and arithmetic
x := 5
y := 10
sum := x + y
print(sum)

// If-else statement
if sum > 10 {
    print("Sum is greater than 10")
} else {
    print("Sum is not greater than 10")
}

// For loop (replaces while)
i := 0
for i < 3 {
    print(i)
    i = i + 1
}

// Function definition and call
add := func(a, b) { 
    return a + b 
}
result := add(5, 3)
print(result)
