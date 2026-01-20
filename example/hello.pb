// A simple hello world test
print("Hello, Pebble!");

// Simple arithmetic
let x = 5 + 3 * 2;
print("5 + 3 * 2 = " + str(x));

// Conditional
if (x > 10) {
    print("x is greater than 10");
} else {
    print("x is not greater than 10");
}

// Simple loop
let i = 0;
while (i < 3) {
    print("Loop count: " + str(i));
    i = i + 1;
}
