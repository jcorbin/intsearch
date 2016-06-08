package word

/*

Package word defines data structures and helper functions for solving
coded-word-addition problems like:

        S E N D
    +   M O R E
    -----------
      M O N E Y

This package contains 3 major pieces:
1. Data type Problem encodes a problem instance, essentially 3 words.
2. Data type PlanProblem manages state and provides methods to for solution
   strategy implementations.
3. Interface SolutionGen decouples the plan strategy from the actual plan
    execution.

*/
