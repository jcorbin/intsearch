def to_number(*digits):
    return sum(d * 10 ** n
               for n, d in enumerate(reversed(digits)))


def choose1(digits, *exclude):
    for digit in digits.difference(exclude):
        yield digit, digits - {digit}

all_digits = set(range(10))

print [
    (to_number(s, e, n, d),
     to_number(m, o, r, e),
     to_number(m, o, n, e, y))

    for d, digits in choose1(all_digits)
    for e, digits in choose1(digits)
    for y, digits in choose1(digits)
    if (d + e) % 10 == y

    for n, digits in choose1(digits)
    for r, digits in choose1(digits)
    if (n + r + (d + e) // 10) % 10 == e

    for o, digits in choose1(digits)
    if (e + o + (n + r + (d + e) // 10) // 10) % 10 == n

    for s, digits in choose1(digits, 0)
    for m, digits in choose1(digits, 0)
    if (s + m + (e + o + (n + r + (d + e) // 10) // 10) // 10) % 10 == o
    if (s + m + (e + o + (n + r + (d + e) // 10) // 10) // 10) // 10 == m]
