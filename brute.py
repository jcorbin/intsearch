def to_number(*digits):
    return sum(d * 10 ** n
               for n, d in enumerate(reversed(digits)))


def choose1(digits, *exclude):
    for digit in digits.difference(exclude):
        yield digit, digits - {digit}

all_digits = set(range(10))

print [
    (send, more, money)
    for s, e, n, d, send, digits in (
        (s, e, n, d, to_number(s, e, n, d), digits)
        for s, digits in choose1(all_digits, 0)
        for e, digits in choose1(digits)
        for n, digits in choose1(digits)
        for d, digits in choose1(digits))
    for m, o, r, more, digits in (
        (m, o, r, to_number(m, o, r, e), digits)
        for m, digits in choose1(digits, 0)
        for o, digits in choose1(digits)
        for r, digits in choose1(digits))
    for y, money, digits in (
        (y, to_number(m, o, n, e, y), digits)
        for y, digits in choose1(digits))
    if send + more == money]
