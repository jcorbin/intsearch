#include <stdio.h>
#include <string.h>
#include <stdbool.h>

#if defined(MEASURE_TIME) || defined(MEASURE_OP_TIME)
#include <time.h>
#endif

#define   EXITCODE_DEAD                    0x01
#define   EXITCODE_CRASH_SEARCH_OVERFLOW   0xfb
#define   EXITCODE_CRASH_STACK_UNDERFLOW   0xfc
#define   EXITCODE_CRASH_STACK_OVERFLOW    0xfd
#define   EXITCODE_CRASH_INVALID_OP        0xfe
#define   EXITCODE_CRASH_INVALID_PI        0xff

#define   OP_JUMP         0x01
#define   OP_JZ           0x02
#define   OP_JNZ          0x03
#define   OP_PUSH         0x04
#define   OP_POP          0x05
#define   OP_DUP          0x06
#define   OP_SWAP         0x07
#define   OP_ADD          0x20
#define   OP_SUB          0x21
#define   OP_MUL          0x22
#define   OP_DIV          0x23
#define   OP_MOD          0x24
#define   OP_LT           0x25
#define   OP_GT           0x26
#define   OP_LTE          0x27
#define   OP_GTE          0x28
#define   OP_INC          0x29
#define   OP_DEC          0x2a
#define   OP_STORE        0x30
#define   OP_LOAD         0x31
#define   OP_IS_SEEN      0x32
#define   OP_SET_SEEN     0x33
#define   OP_FORK         0xfe
#define   OP_EXIT         0xff

#define   MAX_LETTERS   256
#define   MAX_PROGLEN   0x0800
#define   STACK_SIZE    16
#define   SEARCH_SIZE   0x1000


typedef struct ProblemStruct {
    const char *w1, *w2, *w3;
    size_t l1, l2, l3;
    unsigned int base;
    bool known[MAX_LETTERS];
    unsigned char prog[2 * MAX_PROGLEN];
    unsigned char proglen;
} Problem;

typedef struct StateStruct {
    const Problem *prob;
    bool done;
    unsigned char exitcode;
    unsigned int prog_index;
    unsigned int stack_length;
    char letter_map[MAX_LETTERS];
    bool seen[MAX_LETTERS];
    char stack[STACK_SIZE];
} State;

typedef struct StateSpaceStruct {
    unsigned int index;
    unsigned int length;
    State *states;
} StateSpace;

int prog_toString(char *str, size_t n, const unsigned char *prog, const unsigned short pi) {
    const unsigned short i = pi << 1;
    const unsigned short j = i + 1;
    switch (prog[i]) {

        case OP_JUMP:
            return snprintf(str, n, "jump %+i", prog[j]);

        case OP_JZ:
            return snprintf(str, n, "jz %+i", prog[j]);

        case OP_JNZ:
            return snprintf(str, n, "jnz %+i", prog[j]);

        case OP_PUSH:
            return snprintf(str, n, "push %i", prog[j]);

        case OP_POP:
            return snprintf(str, n, "pop");

        case OP_DUP:
            return snprintf(str, n, "dup");

        case OP_SWAP:
            return snprintf(str, n, "swap");

        case OP_ADD:
            return snprintf(str, n, "add");

        case OP_SUB:
            return snprintf(str, n, "sub");

        case OP_MUL:
            return snprintf(str, n, "mul");

        case OP_DIV:
            return snprintf(str, n, "div");

        case OP_MOD:
            return snprintf(str, n, "mod");

        case OP_LT:
            return snprintf(str, n, "lt");

        case OP_GT:
            return snprintf(str, n, "gt");

        case OP_LTE:
            return snprintf(str, n, "lte");

        case OP_GTE:
            return snprintf(str, n, "gte");

        case OP_INC:
            return snprintf(str, n, "inc %i", prog[j]);

        case OP_DEC:
            return snprintf(str, n, "dec %i", prog[j]);

        case OP_STORE:
            return snprintf(str, n, "store %c", prog[j]);

        case OP_LOAD:
            return snprintf(str, n, "load %c", prog[j]);

        case OP_IS_SEEN:
            return snprintf(str, n, "is_seen");

        case OP_SET_SEEN:
            return snprintf(str, n, "set_seen");

        case OP_FORK:
            return snprintf(str, n, "fork");

        case OP_EXIT:
            return snprintf(str, n, "exit %i", prog[j]);

        default:
            return snprintf(str, n, "INVALID %04x %04x", prog[i], prog[j]);
    }
}

int State_stackToString(char *str, size_t n, const State *state) {
    const char *orig = str;

    int i;
    int r;

    r = snprintf(str, n, "[");
    if (r < 0) return r;
    n -= r;
    str += r;

    for (i = 0; i < state->stack_length; ++i) {
        if (i > 0) {
            r = snprintf(str, n, ", %i", state->stack[i]);
        } else {
            r = snprintf(str, n, "%i", state->stack[i]);
        }
        if (r < 0) return r;
        n -= r;
        str += r;
    }

    r = snprintf(str, n, "]");
    if (r < 0) return r;
    n -= r;
    str += r;

    return str - orig;
}

void State_printWords(const State *state) {
    int i;
    struct {const char *w; const size_t l; int i;} words[] = {
        {state->prob->w1, state->prob->l1, 0},
        {state->prob->w2, state->prob->l2, 0},
        {state->prob->w3, state->prob->l3, 0}};

    if (state->prob->l1 > state->prob->l2) {
        words[1].i -= state->prob->l1 - state->prob->l2;
        words[1].i -= state->prob->l3 - state->prob->l1;
        words[0].i -= state->prob->l3 - state->prob->l1;
    } else if (state->prob->l2 > state->prob->l1) {
        words[0].i -= state->prob->l2 - state->prob->l1;
        words[0].i -= state->prob->l3 - state->prob->l2;
        words[1].i -= state->prob->l3 - state->prob->l2;
    } else {
        words[0].i -= state->prob->l3 - state->prob->l2;
        words[1].i -= state->prob->l3 - state->prob->l2;
    }

    for (i = 0; i < 3; ++i) {
        const char *w = words[i].w;
        const size_t l = words[i].l;
        int j = words[i].i;
        printf("  w%i: ", i + 1);
        do {
            if (j < 0) {
                printf("    ");
            } else {
                const char c = w[j];
                const char digit = state->letter_map[c];
                if (digit < 0 || !state->seen[digit]) {
                    printf(" %c:_", c);
                } else {
                    printf(" %c:%i", c, digit);
                }
            }
        } while (++j < l);
        printf("\n");
    }
}

int StateSpace_indexof_state(const StateSpace *space, const State *state) {
    int i;
    State *cur;
    for (
        i = 0, cur = space->states;
        i < space->length;
        i++, cur++
    ) if (cur == state) return i;
    return -1;
}

void StateSpace_printState(const StateSpace *space, const State *state, const unsigned int prog_index) {
    char buf1[256];
    char buf2[256];

    prog_toString(buf1, 256, state->prob->prog, prog_index);
    State_stackToString(buf2, 256, state);

    printf(
        "[%i] %-10s @0x%04x stack=%s\n",
        StateSpace_indexof_state(space, state),
        buf1, prog_index, buf2);
}

State *StateSpace_state_copy(StateSpace *space, State *state) {
    if (space->index >= (space->length - 1)) {
        state->done = true;
        state->exitcode = EXITCODE_CRASH_SEARCH_OVERFLOW;
        return 0L;
    }
    return (State *) memcpy(
        &(space->states[++space->index]),
        state, sizeof(State));
}

void do_op_jump(StateSpace *space, State *state, const char arg) {
    state->prog_index += arg;
}

void do_op_jz(StateSpace *space, State *state, const char arg) {
    if (state->stack_length == 0) {
        state->done = true;
        state->exitcode = EXITCODE_CRASH_STACK_UNDERFLOW;
        return;
    }
    const char c = state->stack[--state->stack_length];
    if (c == 0) {
        state->prog_index += arg;
    } else {
        state->prog_index++;
    }
}

void do_op_jnz(StateSpace *space, State *state, const char arg) {
    if (state->stack_length == 0) {
        state->done = true;
        state->exitcode = EXITCODE_CRASH_STACK_UNDERFLOW;
        return;
    }
    const char c = state->stack[--state->stack_length];
    if (c != 0) {
        state->prog_index += arg;
    } else {
        state->prog_index++;
    }
}

void do_op_push(StateSpace *space, State *state, const char arg) {
    if (state->stack_length >= STACK_SIZE) {
        state->done = true;
        state->exitcode = EXITCODE_CRASH_STACK_OVERFLOW;
        return;
    }
    state->stack[state->stack_length++] = arg;
    state->prog_index++;
}

void do_op_pop(StateSpace *space, State *state, const char arg) {
    if (state->stack_length == 0) {
        state->done = true;
        state->exitcode = EXITCODE_CRASH_STACK_UNDERFLOW;
        return;
    }
    --state->stack_length;
    state->prog_index++;
}

void do_op_store(StateSpace *space, State *state, const char arg) {
    // TODO: consider taking this guard off for perf
    if (state->stack_length == 0) {
        state->done = true;
        state->exitcode = EXITCODE_CRASH_STACK_UNDERFLOW;
        return;
    }
    state->letter_map[arg] = state->stack[state->stack_length - 1];
    state->prog_index++;
}

void do_op_load(StateSpace *space, State *state, const char arg) {
    if (state->stack_length >= STACK_SIZE) {
        state->done = true;
        state->exitcode = EXITCODE_CRASH_STACK_OVERFLOW;
        return;
    }
    state->stack[state->stack_length++] = state->letter_map[arg];
    state->prog_index++;
}

void do_op_is_seen(StateSpace *space, State *state, const char arg) {
    if (state->stack_length == 0) {
        state->done = true;
        state->exitcode = EXITCODE_CRASH_STACK_UNDERFLOW;
        return;
    }
    const unsigned short i = state->stack_length - 1;
    const char digit = state->stack[i];
    state->stack[i] = state->seen[digit] ? 1 : 0;
    state->prog_index++;
}

void do_op_set_seen(StateSpace *space, State *state, const char arg) {
    if (state->stack_length == 0) {
        state->done = true;
        state->exitcode = EXITCODE_CRASH_STACK_UNDERFLOW;
        return;
    }
    const unsigned short i = state->stack_length - 1;
    const char digit = state->stack[i];
    state->stack[i] = state->seen[digit] ? 1 : 0;
    state->seen[digit] = true;
    state->prog_index++;
}

void do_op_dup(StateSpace *space, State *state, const char arg) {
    const unsigned short i = state->stack_length;
    if (i == 0) {
        state->done = true;
        state->exitcode = EXITCODE_CRASH_STACK_UNDERFLOW;
        return;
    }
    if (i >= STACK_SIZE) {
        state->done = true;
        state->exitcode = EXITCODE_CRASH_STACK_OVERFLOW;
        return;
    }
    state->stack[i] = state->stack[i - 1];
    state->stack_length = i + 1;
    state->prog_index++;
}

void do_op_swap(StateSpace *space, State *state, const char arg) {
    if (state->stack_length <= 1) {
        state->done = true;
        state->exitcode = EXITCODE_CRASH_STACK_UNDERFLOW;
        return;
    }
    const unsigned short i = state->stack_length - 1;
    const unsigned short j = i - 1;
    const char tmp = state->stack[i];
    state->stack[i] = state->stack[j];
    state->stack[j] = tmp;
    state->prog_index++;
}

void do_op_add(StateSpace *space, State *state, const char arg) {
    if (state->stack_length <= 1) {
        state->done = true;
        state->exitcode = EXITCODE_CRASH_STACK_UNDERFLOW;
        return;
    }
    const unsigned short i = --state->stack_length;
    const unsigned short j = i - 1;
    const char b = state->stack[i];
    const char a = state->stack[j];
    state->stack[j] = a + b;
    state->prog_index++;
}

void do_op_sub(StateSpace *space, State *state, const char arg) {
    if (state->stack_length <= 1) {
        state->done = true;
        state->exitcode = EXITCODE_CRASH_STACK_UNDERFLOW;
        return;
    }
    const unsigned short i = --state->stack_length;
    const unsigned short j = i - 1;
    const char b = state->stack[i];
    const char a = state->stack[j];
    state->stack[j] = a - b;
    state->prog_index++;
}

void do_op_mul(StateSpace *space, State *state, const char arg) {
    if (state->stack_length <= 1) {
        state->done = true;
        state->exitcode = EXITCODE_CRASH_STACK_UNDERFLOW;
        return;
    }
    const unsigned short i = --state->stack_length;
    const unsigned short j = i - 1;
    const char b = state->stack[i];
    const char a = state->stack[j];
    state->stack[j] = a * b;
    state->prog_index++;
}

void do_op_div(StateSpace *space, State *state, const char arg) {
    if (state->stack_length <= 1) {
        state->done = true;
        state->exitcode = EXITCODE_CRASH_STACK_UNDERFLOW;
        return;
    }
    const unsigned short i = --state->stack_length;
    const unsigned short j = i - 1;
    const char b = state->stack[i];
    const char a = state->stack[j];
    state->stack[j] = a / b;
    state->prog_index++;
}

void do_op_mod(StateSpace *space, State *state, const char arg) {
    if (state->stack_length <= 1) {
        state->done = true;
        state->exitcode = EXITCODE_CRASH_STACK_UNDERFLOW;
        return;
    }
    const unsigned short i = --state->stack_length;
    const unsigned short j = i - 1;
    const char b = state->stack[i];
    const char a = state->stack[j];
    state->stack[j] = a % b;
    state->prog_index++;
}

void do_op_lt(StateSpace *space, State *state, const char arg) {
    if (state->stack_length <= 1) {
        state->done = true;
        state->exitcode = EXITCODE_CRASH_STACK_UNDERFLOW;
        return;
    }
    const unsigned short i = --state->stack_length;
    const unsigned short j = i - 1;
    const char b = state->stack[i];
    const char a = state->stack[j];
    state->stack[j] = a < b ? 1 : 0;
    state->prog_index++;
}

void do_op_gt(StateSpace *space, State *state, const char arg) {
    if (state->stack_length <= 1) {
        state->done = true;
        state->exitcode = EXITCODE_CRASH_STACK_UNDERFLOW;
        return;
    }
    const unsigned short i = --state->stack_length;
    const unsigned short j = i - 1;
    const char b = state->stack[i];
    const char a = state->stack[j];
    state->stack[j] = a > b ? 1 : 0;
    state->prog_index++;
}

void do_op_lte(StateSpace *space, State *state, const char arg) {
    if (state->stack_length <= 1) {
        state->done = true;
        state->exitcode = EXITCODE_CRASH_STACK_UNDERFLOW;
        return;
    }
    const unsigned short i = --state->stack_length;
    const unsigned short j = i - 1;
    const char b = state->stack[i];
    const char a = state->stack[j];
    state->stack[j] = a <= b ? 1 : 0;
    state->prog_index++;
}

void do_op_gte(StateSpace *space, State *state, const char arg) {
    if (state->stack_length <= 1) {
        state->done = true;
        state->exitcode = EXITCODE_CRASH_STACK_UNDERFLOW;
        return;
    }
    const unsigned short i = --state->stack_length;
    const unsigned short j = i - 1;
    const char b = state->stack[i];
    const char a = state->stack[j];
    state->stack[j] = a >= b ? 1 : 0;
    state->prog_index++;
}

void do_op_inc(StateSpace *space, State *state, const char arg) {
    if (state->stack_length == 0) {
        state->done = true;
        state->exitcode = EXITCODE_CRASH_STACK_UNDERFLOW;
        return;
    }
    state->stack[state->stack_length - 1] += arg;
    state->prog_index++;
}

void do_op_dec(StateSpace *space, State *state, const char arg) {
    if (state->stack_length == 0) {
        state->done = true;
        state->exitcode = EXITCODE_CRASH_STACK_UNDERFLOW;
        return;
    }
    state->stack[state->stack_length - 1] -= arg;
    state->prog_index++;
}

void do_op_fork(StateSpace *space, State *state, const char arg) {
    if (space->index >= (space->length - arg)) {
        state->done = true;
        state->exitcode = EXITCODE_CRASH_SEARCH_OVERFLOW;
        return;
    }

    if (state->stack_length >= STACK_SIZE) {
        state->done = true;
        state->exitcode = EXITCODE_CRASH_STACK_OVERFLOW;
        return;
    }
    const unsigned int i = state->stack_length++;
    state->stack[i] = 0;
    state->prog_index++;

    State *next_state = &(space->states[space->index]);
    space->index += arg;
    unsigned int j = 1;
    for (; j <= arg; ++j) {
        memcpy(++next_state, state, sizeof(State));
        next_state->stack[i] = j;
    }
}

void do_op_exit(StateSpace *space, State *state, const char arg) {
    state->done = true;
    state->exitcode = arg;
}

void do_op_invalid(StateSpace *space, State *state, const char arg) {
    state->done = true;
    state->exitcode = EXITCODE_CRASH_INVALID_OP;
}

void StateSpace_state_tick(StateSpace *space, State *state) {
    const unsigned short i = state->prog_index;
    if (i >= state->prob->proglen) {
        state->done = true;
        state->exitcode = EXITCODE_CRASH_INVALID_PI;
        return;
    }

    const unsigned short j = i << 1;
    const unsigned char *prog = &(state->prob->prog[0]);
    const unsigned char op = prog[j];
    const unsigned char arg = prog[j + 1];

#ifdef MEASURE_OP_TIME
    clock_t op_start = clock();
#endif

    switch (op) {

        case OP_JUMP:
            do_op_jump(space, state, arg);
            break;

        case OP_JZ:
            do_op_jz(space, state, arg);
            break;

        case OP_JNZ:
            do_op_jnz(space, state, arg);
            break;

        case OP_PUSH:
            do_op_push(space, state, arg);
            break;

        case OP_POP:
            do_op_pop(space, state, arg);
            break;

        case OP_DUP:
            do_op_dup(space, state, arg);
            break;

        case OP_SWAP:
            do_op_swap(space, state, arg);
            break;

        case OP_ADD:
            do_op_add(space, state, arg);
            break;

        case OP_SUB:
            do_op_sub(space, state, arg);
            break;

        case OP_MUL:
            do_op_mul(space, state, arg);
            break;

        case OP_DIV:
            do_op_div(space, state, arg);
            break;

        case OP_MOD:
            do_op_mod(space, state, arg);
            break;

        case OP_LT:
            do_op_lt(space, state, arg);
            break;

        case OP_GT:
            do_op_gt(space, state, arg);
            break;

        case OP_LTE:
            do_op_lte(space, state, arg);
            break;

        case OP_GTE:
            do_op_gte(space, state, arg);
            break;

        case OP_INC:
            do_op_inc(space, state, arg);
            break;

        case OP_DEC:
            do_op_dec(space, state, arg);
            break;

        case OP_STORE:
            do_op_store(space, state, arg);
            break;

        case OP_LOAD:
            do_op_load(space, state, arg);
            break;

        case OP_IS_SEEN:
            do_op_is_seen(space, state, arg);
            break;

        case OP_SET_SEEN:
            do_op_set_seen(space, state, arg);
            break;

        case OP_FORK:
            do_op_fork(space, state, arg);
            break;

        case OP_EXIT:
            do_op_exit(space, state, arg);
            break;

        default:
            do_op_invalid(space, state, arg);
            break;
    }

#ifdef MEASURE_OP_TIME
    clock_t op_end = clock();
#endif

#ifdef PRINT_TRACE
    StateSpace_printState(space, state, i);

#ifdef MEASURE_OP_TIME
    printf("  op_time: clocks=%li ns=%i\n",
           op_end - op_start,
           (int)((double)(op_end - op_start) * 1e9 / CLOCKS_PER_SEC));
#endif

    if (op == OP_STORE) {
        State_printWords(state);
    }
#endif
}

void Problem_push_op(Problem *prob, unsigned char op, unsigned char arg) {
    // TODO: guard prob->proglen < MAX_PROGLEN
    const unsigned short i = prob->proglen++ << 1;
    prob->prog[i] = op;
    prob->prog[i + 1] = arg;
}

void Problem_fix(Problem *prob, const char c, const char digit, const bool check_seen) {
#ifdef PRINT_PLAN
    printf("  - fix %c = %i (%s)\n",
            c, digit,
            check_seen ? "check" : "no check");
#endif
    Problem_push_op(prob, OP_PUSH, digit);             // [..., digit]
    Problem_push_op(prob, OP_DUP, digit);              // [..., digit, digit]
    Problem_push_op(prob, OP_SET_SEEN, 0);             // [..., digit, was_seen]
    if (check_seen) {                                  // ...
        Problem_push_op(prob, OP_JZ, 2);               // [..., digit]
        Problem_push_op(prob, OP_EXIT, EXITCODE_DEAD); // ...
    } else {                                           // ...
        Problem_push_op(prob, OP_POP, c);              // [..., digit]
    }                                                  // ...
    Problem_push_op(prob, OP_STORE, c);                // ...
    prob->known[c] = true;                             // ...
#ifdef PRINT_PLAN
    printf("    - known %c\n", c);
#endif
}

void Problem_choose_dfs(Problem *prob, const char c) {
#ifdef PRINT_PLAN
    printf("    - choose dfs %c\n", c);
#endif
    /* for i=0; i<base; ++i
     *   if !fork continue
     *   if set_seen i continue
     *   store letter, i
     *   break
     */

    bool is_first =
        prob->w1[0] == c ||
        prob->w2[0] == c ||
        prob->w3[0] == c;
    unsigned char initial = is_first ? 1 : 0;

    Problem_push_op(prob, OP_PUSH, initial);         // 0  // [..., 0]           // i = 0
    Problem_push_op(prob,   OP_FORK, 1);             // 1  // [..., i, is_child] // fork
    Problem_push_op(prob,   OP_JZ, 7);               // 2  // [..., i]           // continue if not child
    Problem_push_op(prob,   OP_DUP, 0);              // 3  // [..., i, i]        // loop start; dup for seen
    Problem_push_op(prob,   OP_SET_SEEN, 0);         // 4  // [..., i, was_seen] // was_seen i
    Problem_push_op(prob,   OP_JZ, 2);               // 5  // [..., i]           // exit if seen already
    Problem_push_op(prob,   OP_EXIT, EXITCODE_DEAD); // 6  // [..., i]           // exit if seen already
    Problem_push_op(prob,   OP_STORE, c);            // 7  // [..., i]           // store letter = i
    Problem_push_op(prob,   OP_JUMP, 7);             // 8  // [..., i]           // break
    Problem_push_op(prob, OP_INC, 1);                // 9  // [..., ++i]         // ++i
    Problem_push_op(prob, OP_DUP, 0);                // 10 // [..., i, i]        // dup for cmp
    Problem_push_op(prob, OP_PUSH, prob->base);      // 11 // [..., i, i, base]  // push for cmp
    Problem_push_op(prob, OP_LT, 0);                 // 12 // [..., i, i < base] // i < base
    Problem_push_op(prob, OP_JNZ, -12);              // 13 // [..., i]           // loop check
    Problem_push_op(prob, OP_EXIT, EXITCODE_DEAD);   // 14 // [..., i]           // parent exit

    prob->known[c] = true;
#ifdef PRINT_PLAN
    printf("    - known %c\n", c);
#endif
}

void Problem_choose_bfs(Problem *prob, const char c) {
#ifdef PRINT_PLAN
    printf("    - choose bfs %c\n", c);
#endif

    /* // not is_first
     * if !fork base exit dead
     * --N
     * if set_seen N exit dead
     * store letter, N
     *
     * // is_first
     * if !fork base-1 exit dead
     * if set_seen N exit dead
     * store letter, N
     */

    bool is_first =
        prob->w1[0] == c ||
        prob->w2[0] == c ||
        prob->w3[0] == c;
    unsigned char forks = is_first ? prob->base - 1 : prob->base;

    Problem_push_op(prob, OP_FORK, forks);         // [..., N]
    Problem_push_op(prob, OP_DUP, 0);              // [..., N, N]
    Problem_push_op(prob, OP_JNZ, 2);              // [..., N]
    Problem_push_op(prob, OP_EXIT, EXITCODE_DEAD); // ...
    if (!is_first) {                               // ...
        Problem_push_op(prob, OP_DEC,  1);         // [..., --N]
    }                                              // ...
    Problem_push_op(prob, OP_DUP, 0);              // [..., N, N]
    Problem_push_op(prob, OP_SET_SEEN, 0);         // [..., N, was_seen]
    Problem_push_op(prob, OP_JZ, 2);               // [..., N]
    Problem_push_op(prob, OP_EXIT, EXITCODE_DEAD); // ...
    Problem_push_op(prob, OP_STORE, c);            // [..., N]

    prob->known[c] = true;
#ifdef PRINT_PLAN
    printf("    - known %c\n", c);
#endif
}

void Problem_load_or_choose(Problem *prob, const char c) {
    if (prob->known[c]) {
#ifdef PRINT_PLAN
        printf("    - load %c\n", c);
#endif
        Problem_push_op(prob, OP_LOAD, c);
        return;
    }
    /* Problem_choose_dfs(prob, c); */
    Problem_choose_bfs(prob, c);
}

void Problem_solve_sum(Problem *prob, const char c1, const char c2, const char c3) {                                                // [carry]
#ifdef PRINT_PLAN
    printf("  - solve %c + %c = %c for %c\n", c1, c2, c3, c3);
#endif
    Problem_load_or_choose(prob, c1);              // [carry, c1]
    Problem_push_op(prob, OP_ADD, 0);              // [carry + c1]
    Problem_load_or_choose(prob, c2);              // [carry + c1, c2]
    Problem_push_op(prob, OP_ADD, 0);              // [carry + c1 + c2]
    Problem_push_op(prob, OP_DUP, 0);              // [carry + c1 + c2, carry + c1 + c2]
    Problem_push_op(prob, OP_PUSH, prob->base);    // [carry + c1 + c2, carry + c1 + c2, base]
    Problem_push_op(prob, OP_MOD, 0);              // [carry + c1 + c2, (carry + c1 + c2) % base]
    Problem_push_op(prob, OP_DUP, 0);              // [carry + c1 + c2, c3, c3]
    Problem_push_op(prob, OP_SET_SEEN, 0);         // [carry + c1 + c2, c3, was_seen]
    Problem_push_op(prob, OP_JZ, 2);               // [carry + c1 + c2, c3]
    Problem_push_op(prob, OP_EXIT, EXITCODE_DEAD); // ...
    Problem_push_op(prob, OP_STORE, c3);           // [carry + c1 + c2, c3]
    Problem_push_op(prob, OP_POP, 0);              // [carry + c1 + c2]
    Problem_push_op(prob, OP_PUSH, prob->base);    // [carry + c1 + c2,  base]
    Problem_push_op(prob, OP_DIV, 0);              // [(carry + c1 + c2) / base]
    prob->known[c3] = true;                        // [carry]
#ifdef PRINT_PLAN
    printf("    - known %c\n", c3);
#endif
}

// solve for c1
// carry + c1 + c2 = c3 (mod base)
// carry + c2 - c3 = -c1 (mod base)
// carry + c2 - c3 = base - c1
// c1 = base - (carry + c2 - c3)

void Problem_solve_summand(Problem *prob, const char c1, const char c2, const char c3) {                                                // [carry]
#ifdef PRINT_PLAN
    printf("  - solve %c + %c = %c for %c\n", c1, c2, c3, c1);
#endif
    Problem_load_or_choose(prob, c2);              // [carry, c2]
    Problem_push_op(prob, OP_ADD, 0);              // [carry + c2]
    Problem_push_op(prob, OP_DUP, 0);              // [carry + c2, carry + c2]
    Problem_push_op(prob, OP_LOAD, c3);            // [carry + c2, carry + c2, c3]
    Problem_push_op(prob, OP_SUB, 0);              // [carry + c2, carry + c2 - c3]
    Problem_push_op(prob, OP_PUSH, prob->base);    // [carry + c2, carry + c2 - c3, base]
    Problem_push_op(prob, OP_SWAP, 0);             // [carry + c2, base, carry + c2 - c3]
    Problem_push_op(prob, OP_SUB, 0);              // [carry + c2, base - (carry + c2 - c3)]
    Problem_push_op(prob, OP_PUSH, prob->base);    // [carry + c2, base - (carry + c2 - c3), base]
    Problem_push_op(prob, OP_MOD, 0);              // [carry + c2, c1]
    Problem_push_op(prob, OP_DUP, 0);              // [carry + c2, c1, c1]
    Problem_push_op(prob, OP_SET_SEEN, 0);         // [carry + c2, c1, was_seen]
    Problem_push_op(prob, OP_JZ, 2);               // [carry + c2, c1]
    Problem_push_op(prob, OP_EXIT, EXITCODE_DEAD); // ...
    Problem_push_op(prob, OP_STORE, c1);           // [carry + c2, c1]
    Problem_push_op(prob, OP_ADD, 0);              // [carry + c2 + c1]
    Problem_push_op(prob, OP_PUSH, prob->base);    // [carry + c2 + c1, base]
    Problem_push_op(prob, OP_DIV, 0);              // [(carry + c2 + c1) / base]
    prob->known[c1] = true;                        // [carry]
#ifdef PRINT_PLAN
    printf("    - known %c\n", c1);
#endif
}

void Problem_check_sum(Problem *prob, const char c1, const char c2, const char c3) {                                                // [carry]
#ifdef PRINT_PLAN
    printf("  - check %c + %c = %c\n", c1, c2, c3);
#endif
    Problem_push_op(prob, OP_LOAD, c1);            // [carry, c1]
    Problem_push_op(prob, OP_ADD, 0);              // [carry + c1]
    Problem_push_op(prob, OP_LOAD, c2);            // [carry + c1, c2]
    Problem_push_op(prob, OP_ADD, 0);              // [carry + c1 + c2]
    Problem_push_op(prob, OP_DUP, 0);              // [carry + c1 + c2, carry + c1 + c2]
    Problem_push_op(prob, OP_PUSH, prob->base);    // [carry + c1 + c2, carry + c1 + c2, base]
    Problem_push_op(prob, OP_MOD, 0);              // [carry + c1 + c2, (carry + c1 + c2) % base]
    Problem_push_op(prob, OP_LOAD, c3);            // [carry + c1 + c2, (carry + c1 + c2) % base, c3]
    Problem_push_op(prob, OP_SUB, 0);              // [carry + c1 + c2, cmp]
    Problem_push_op(prob, OP_JZ, 2);               // [carry + c1 + c2]
    Problem_push_op(prob, OP_EXIT, EXITCODE_DEAD); // ...
    Problem_push_op(prob, OP_PUSH, prob->base);    // [carry + c1 + c2, base]
    Problem_push_op(prob, OP_DIV, 0);              // [(carry + c1 + c2) / base]
}                                                  // [carry]

void Problem_check_final(Problem *prob, const char c1, const char c2, const char c3) {                                                // [carry]
#ifdef PRINT_PLAN
    printf("  - load %c\n", c3);
#endif
    Problem_push_op(prob, OP_LOAD, c3);            // [carry, c3]
    if (c1 != 0x00) {                              //
        Problem_load_or_choose(prob, c1);          // [carry, c3, c1]
        Problem_push_op(prob, OP_ADD, 0);          // [carry, c3 + c1]
    } else if (c2 != 0x00) {                       //
        Problem_load_or_choose(prob, c1);          // [carry, c3, c2]
        Problem_push_op(prob, OP_ADD, 0);          // [carry, c3 + c2]
    }                                              // [carry, X]
#ifdef PRINT_PLAN
    printf("  - check final\n");
#endif
    Problem_push_op(prob, OP_SUB, 0);              // [cmp]
    Problem_push_op(prob, OP_JZ, 2);               // []
    Problem_push_op(prob, OP_EXIT, EXITCODE_DEAD); // ...
    Problem_push_op(prob, OP_EXIT, 0);             // ...
}

void Problem_compile(Problem *prob) {
#ifdef PRINT_PLAN
    printf("plan:\n");
#endif

    size_t
        i1 = prob->l1,
        i2 = prob->l2,
        i3 = prob->l3;

    /* if we have
     *      ABC...
     *   +  DEF...
     *   ---------
     *     GHI....
     * Then G must = 1 since its only summand is carry
     */
    if (i3 > i2 && i3 > i1) {
        Problem_fix(prob, prob->w3[0], 1, false);
    }

    // initial carry
    Problem_push_op(prob, OP_PUSH, 0);

    // solve each column
    while (i1 > 0 && i2 > 0 && i3 > 0) {
        const char
            c1 = prob->w1[--i1],
            c2 = prob->w2[--i2],
            c3 = prob->w3[--i3];
        if (!prob->known[c3]) {
            Problem_solve_sum(prob, c1, c2, c3);
        } else if (!prob->known[c1]) {
            Problem_solve_summand(prob, c1, c2, c3);
        } else if (!prob->known[c2]) {
            Problem_solve_summand(prob, c2, c1, c3);
        } else {
            Problem_check_sum(prob, c1, c2, c3);
        }
    }

    // verify any final partial column
    if (i3 > 0) {
        Problem_check_final(prob,
            i1 > 0 ? prob->w1[--i1] : 0x00,
            i2 > 0 ? prob->w2[--i2] : 0x00,
                     prob->w3[--i3]);
    } else {
        Problem_push_op(prob, OP_EXIT, 0);
    }
#ifdef PRINT_PLAN
    printf("\n");
#endif
}

int Problem_setup(Problem *prob, const char *w1, const char *w2, const char *w3) {
    memset(prob, 0, sizeof(Problem));

    prob->w1 = w1;
    prob->l1 = strlen(w1);
#ifdef PRINT_PLAN
    printf("w1: %s\n", prob->w1);
#endif

    prob->w2 = w2;
    prob->l2 = strlen(w2);
#ifdef PRINT_PLAN
    printf("w2: %s\n", prob->w2);
#endif

    prob->w3 = w3;
    prob->l3 = strlen(w3);
#ifdef PRINT_PLAN
    printf("w3: %s\n", prob->w3);
#endif

    prob->base = 10;
#ifdef PRINT_PLAN
    printf("base: %i\n\n", prob->base);
#endif

    if (prob->l3 < prob->l2 ||
        prob->l3 < prob->l1) {
        return -1;
    }

    if (prob->l3 - prob->l1 > 1 &&
        prob->l3 - prob->l1 > 1) {
        return -1;
    }

    Problem_compile(prob);

#ifdef PRINT_PLAN
    unsigned short i;
    char buf[255];
    printf("program:\n");
    printf("  - %i instructions\n", prob->proglen);
    for (i = 0; i < prob->proglen; ++i) {
        prog_toString(buf, 256, prob->prog, i);
        printf("  0x%04x: %s\n", i, buf);
    }
    printf("\n");
#endif

    return 0;
}

int main(const int argc, const char *argv[]) {
    if (argc != 4) {
        return 1;
    }

    Problem prob;

#ifdef MEASURE_TIME
    clock_t begin = clock();
#endif
    if (Problem_setup(&prob, argv[1], argv[2], argv[3]) != 0) {
        return 2;
    }

#ifdef MEASURE_TIME
    clock_t setup_done = clock();
    printf("time_setup:\n");
    printf("  %li clocks\n", setup_done - begin);
    printf("  %.1fµs\n\n", (double)(setup_done - begin) * 1e6 / CLOCKS_PER_SEC);
    setup_done = clock();
#endif

    StateSpace search;
    State states[SEARCH_SIZE];
    search.index = 0;
    search.length = SEARCH_SIZE;
    search.states = &(states[0]);

    State *state = search.states;
    memset(state, 0, sizeof(State));
    state->prob = &prob;

    unsigned int i = 0;
    while (i < MAX_LETTERS) state->letter_map[i++] = -1;

#ifdef MEASURE_TIME
    clock_t run_start = clock();
    printf("time_alloc:\n");
    printf("  %li clocks\n", run_start - setup_done);
    printf("  %.1fµs\n\n", (double)(run_start - setup_done) * 1e6 / CLOCKS_PER_SEC);
    run_start = clock();
    clock_t end;
#endif

    do {
        StateSpace_state_tick(&search, state);
        state = &(search.states[search.index]);
        while (state->done) {
            if (state->exitcode == 0) {
#ifdef MEASURE_TIME
                end = clock();
                printf("time_run:\n");
                printf("  %li clocks\n", end - run_start);
                printf("  %.1fµs\n", (double)(end - run_start) * 1e6 / CLOCKS_PER_SEC);
#endif
                printf("\nfound\n");
                State_printWords(state);
                return 0;
            }
            if (search.index == 0) {
#ifdef MEASURE_TIME
                end = clock();
                printf("time_run:\n");
                printf("  %li clocks\n", end - run_start);
                printf("  %.1fµs\n", (double)(end - run_start) * 1e6 / CLOCKS_PER_SEC);
#endif
                printf("\nno result\n");
                return 3;
            }
            state = &(search.states[--search.index]);
        }
    } while (true);
}
