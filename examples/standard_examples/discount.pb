// calculate_discount returns the discount amount for an order
func calculate_discount(order_total, is_member) {
    discount := 0

    if order_total > 100 {
        discount = order_total * 0.1
    }

    if is_member {
        discount = discount + (order_total * 0.05)
    }

    return discount
}

// Example usage
total := 150
member := true
res := calculate_discount(total, member)
print("Total Discount: ", res)
