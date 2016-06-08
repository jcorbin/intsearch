package runnable

/*

Package runnable provides an implementation of word.SolutionGen based on a
solution state structure and a runnable step interface that operates on
solution state.  Runnable steps may mutate the state passed to them, and may
also create one or more copies of the current state for further
non-deterministic exploration.  A simple search facility is then provided over
solution state space.

*/
