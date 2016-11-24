# Description

This repository hold various solutions to the search problem:

    Find a mapping of decimal digits to the letters in the following:

            S E N D
        +   M O R E
        -----------
          M O N E Y

    Such that the addition works.

# Inspiration

I came across search-based solutions to this problem from
http://blog.plover.com/prog/monad-search-2.html.

# Solutions

- the `python_static_comprehension_2015-08` branch contains a Python solution
  using simple comprehensions for brute force search
- the `js_progsearch_2015-08` branch contains a Node.js solution using an
  object-oriented `ProgSearch` approach
- the `c_stack_machine_2015-11` branch contains a C solution using a custom
  virtual stack machine
- the `go_2016-04` branch contains a Go solution that also uses a programmatic
  search approach similar to the C solution
