'use strict';

var util = require('util');
var assert = require('assert');

var Operations = {};

Operations.chooseLetter = function chooseLetter(state, op) {
    var nexts = [];
    var start = op.isInitial ? 1 : 0;

    for (var digit = start; digit < op.base; digit++) {
        if (!state.chosen[digit]) {
            var next = (state.alloc()).copyFrom(state);
            next.chosen[digit] = true;
            next.values[op.letter] = digit;
            nexts.push(next);
        }
    }

    return nexts;
};

Operations.result = function result(state, op) {
    var res = new Array(op.values.length);
    for (var i = 0; i < op.values.length; i++) {
        res[i] = state.values[op.values[i]];
    }
    state.result = res;
};

Operations.sum = function sum(state, op) {
    var res = 0;

    for (var i = 0; i < op.values.length; i++) {
        var letter = op.values[i];
        var digit = state.values[letter];
        res += digit;
    }
    state.values[op.store] = res;
};

Operations.remainder = function remainder(state, op) {
    var dividend = state.values[op.dividend];
    state.values[op.store] = dividend % op.divisor;
};

Operations.floordiv = function floordiv(state, op) {
    var dividend = state.values[op.dividend];
    state.values[op.store] = Math.floor(dividend / op.divisor);
};

Operations.equal = function equal(state, op) {
    if (state.valid) {
        var arg1 = state.values[op.arg1];
        var arg2 = state.values[op.arg2];
        state.valid = arg1 === arg2;
    }
};

Operations.toNumber = function toNumber(state, op) {
    var value = 0;
    // for (var i = op.values.length-1; i >= 0; i--)
    for (var i = 0; i < op.values.length; i++) {
        value *= op.base;
        value += state.values[op.values[i]];
    }
    state.values[op.store] = value;
};

function letterValuesFrom(words) {
    var values = {};
    words.forEach(function each(word) {
        for (var i = 0; i < word.length; i++) {
            if (values[word[i]] === undefined) {
                values[word[i]] = null;
            }
        }
    });
    return values;
}

function baseFor(n) {
    if (n < 10) {
        return 10;
    } else if (n < 16) {
        return 16;
    } else if (n < 32) {
        return 32;
    } else if (n < 64) {
        return 64;
    } else {
        assert(false, 'not supported');
    }
}

function compileWordProblem(word1, word2, word3) {
    var lenDiff = word3.length - word2.length;
    assert(word1.length <= 8);
    assert(word1.length === word2.length);
    assert(lenDiff === 0 || lenDiff === 1);

    var values = letterValuesFrom([word1, word2, word3]);
    var letters = Object.keys(values);
    var base = baseFor(letters.length);

    var plan = [];

    var initialState = new WordProblemState();
    initialState.chosen.length = base;
    for (var i = 0; i < base; i++) {
        initialState.chosen[i] = false;
    }
    initialState.pi = 1;

    plan.push({
        state: initialState
    });

    var seen = {};

    var lastCarry = null;
    for (var i = 1; i <= word1.length; i++) {
        var quo = 'quo' + i;

        var let1 = addLetter(word1, word1.length - i);
        var let2 = addLetter(word2, word2.length - i);
        var let3 = addLetter(word3, word3.length - i);

        plan.push({
            op: Operations.sum,
            store: 'sum',
            values: lastCarry ? [lastCarry, let1, let2] : [let1, let2],
        });
        plan.push({
            op: Operations.remainder,
            store: 'rem',
            dividend: 'sum',
            divisor: base
        });
        plan.push({
            op: Operations.floordiv,
            store: quo,
            dividend: 'sum',
            divisor: base
        });
        plan.push({
            op: Operations.equal,
            arg1: 'rem',
            arg2: let3
        });
        lastCarry = quo;
    }

    plan.push({
        op: Operations.equal,
        arg1: lastCarry,
        arg2: lenDiff ? addLetter(word3, 0) : 0
    });

    plan.push({
        op: Operations.toNumber,
        store: 'word1',
        values: word1.split(''),
        base: base
    });

    plan.push({
        op: Operations.toNumber,
        store: 'word2',
        values: word2.split(''),
        base: base
    });

    plan.push({
        op: Operations.toNumber,
        store: 'word3',
        values: word3.split(''),
        base: base
    });

    plan.push({
        op: Operations.result,
        values: ['word1', 'word2', 'word3']
    });

    return plan;

    function addLetter(word, i) {
        var letter = word[i];
        if (!seen[letter]) {
            seen[letter] = true;
            plan.push({
                op: Operations.chooseLetter,
                letter: letter,
                base: base,
                isInitial: i === 0
            });
        }
        return letter;
    }
}

function WordProblemValues() {
    var self = this;

    self.a = null;
    self.b = null;
    self.c = null;
    self.d = null;
    self.e = null;
    self.f = null;
    self.g = null;
    self.h = null;
    self.i = null;
    self.j = null;
    self.k = null;
    self.l = null;
    self.m = null;
    self.n = null;
    self.o = null;
    self.p = null;
    self.q = null;
    self.r = null;
    self.s = null;
    self.t = null;
    self.u = null;
    self.v = null;
    self.w = null;
    self.x = null;
    self.y = null;
    self.z = null;

    self.sum = null;
    self.rem = null;

    self.quo1 = null;
    self.quo2 = null;
    self.quo3 = null;
    self.quo4 = null;
    self.quo5 = null;
    self.quo6 = null;
    self.quo7 = null;
    self.quo8 = null;
    self.quo9 = null;

    self.word1 = null;
    self.word2 = null;
    self.word3 = null;
}

WordProblemValues.prototype.copyFrom = function copyFrom(other) {
    var self = this;

    self.a = other.a;
    self.b = other.b;
    self.c = other.c;
    self.d = other.d;
    self.e = other.e;
    self.f = other.f;
    self.g = other.g;
    self.h = other.h;
    self.i = other.i;
    self.j = other.j;
    self.k = other.k;
    self.l = other.l;
    self.m = other.m;
    self.n = other.n;
    self.o = other.o;
    self.p = other.p;
    self.q = other.q;
    self.r = other.r;
    self.s = other.s;
    self.t = other.t;
    self.u = other.u;
    self.v = other.v;
    self.w = other.w;
    self.x = other.x;
    self.y = other.y;
    self.z = other.z;

    self.sum = other.sum;
    self.rem = other.rem;

    self.quo1 = other.quo1;
    self.quo2 = other.quo2;
    self.quo3 = other.quo3;
    self.quo4 = other.quo4;
    self.quo5 = other.quo5;
    self.quo6 = other.quo6;
    self.quo7 = other.quo7;
    self.quo8 = other.quo8;
    self.quo9 = other.quo9;

    self.word1 = other.word1;
    self.word2 = other.word2;
    self.word3 = other.word3;
};

function WordProblemState() {
    var self = this;

    self.pi = 0;
    self.valid = true;
    self.result = null;
    self.chosen = [];
    self.values = new WordProblemValues();
}

WordProblemState.prototype.alloc = function alloc() {
    return new WordProblemState();
};

WordProblemState.prototype.copyFrom = function copyFrom(state) {
    var self = this;

    self.pi = state.pi;
    self.valid = true;
    self.result = null;

    self.chosen.length = state.chosen.length;
    for (var i = 0; i < state.chosen.length; i++) {
        self.chosen[i] = state.chosen[i];
    }
    self.values.copyFrom(state.values);

    return self;
};

function ProgSearch(stateType) {
    var self = this;

    self.stateType = stateType;
    self.freelist = [];
    self.frontier = [];
    self.executed = 0;
    self.expanded = 1;
    self.dirty = false;
}

ProgSearch.prototype.free = function free(state) {
    var self = this;

    self.freelist.push(state);
};

ProgSearch.prototype.makeNewState = function makeNewState() {
    var self = this;

    var state = new self.stateType();
    state.alloc = alloc;
    return state;

    function alloc() {
        return self.alloc();
    }
};

ProgSearch.prototype.alloc = function alloc() {
    var self = this;

    if (self.freelist.length) {
        return self.freelist.shift();
    } else {
        return self.makeNewState();
    }
};

ProgSearch.prototype.run = function run(plan, each) {
    var self = this;

    self.executed = 0;
    self.expanded = 1;
    var state = self.alloc().copyFrom(plan[0].state);
    self.frontier.push(state);

    while (self.frontier.length) {
        state = self.frontier.shift();
        self.expand(plan, state);
        if (state.valid && state.result) {
            if (each(state)) {
                break;
            }
        }
        self.free(state);
    }

    while (self.frontier.length) {
        self.free(self.frontier.shift());
    }
};

ProgSearch.prototype.expand = function expand(plan, state) {
    var self = this;

    var succ = self.execute(plan, state);
    if (succ) {
        self.expanded += succ.length;
        self.frontier.push.apply(self.frontier, succ);
        self.heapify();
    }
};

ProgSearch.prototype.execute = function execute(plan, state) {
    var self = this;

    var succ = null;

    while (!succ && state.valid && state.pi < plan.length) {
        self.executed++;
        var op = plan[state.pi++];
        succ = op.op(state, op);
    }

    if (state.valid && state.result === null) {
        state.valid = false;
    }

    return succ;
};

ProgSearch.prototype.heapify = function heapify() {
    var self = this;

    for (var i = Math.floor(self.frontier.length / 2 - 1); i >= 0; i--) {
        self.siftdown(i);
    }
};

ProgSearch.prototype.swap = function swap(i, j) {
    var self = this;

    var a = self.frontier[i];
    var b = self.frontier[j];
    self.frontier[i] = b;
    self.frontier[j] = a;

    return a;
};

ProgSearch.prototype.siftup = function siftup(i) {
    var self = this;

    while (i > 0) {
        var j = Math.floor((i - 1) / 2);
        if (self.frontier[i].pi > self.frontier[j].pi) {
            self.swap(i, j);
            i = j;
        }
    }

    return i;
};

ProgSearch.prototype.siftdown = function siftdown(i) {
    var self = this;

    while (true) {
        var left = (2 * i) + 1;
        if (left >= self.frontier.length) {
            return;
        }

        var right = left + 1;
        var child = left;
        if (right < self.frontier.length &&
            self.frontier[right].pi > self.frontier[left].pi) {
            child = right;
        }

        if (self.frontier[child].pi <= self.frontier[i].pi) {
            return;
        }

        self.swap(child, i);
        i = child;

    }
    return i;
};

function hrtimeDiff(a, b) {
    var r = [
        a[0] - b[0],
        a[1] - b[1]
    ];
    if (r[1] < 0) {
        r[0]--;
        r[1] += 1e9;
    }
    return r;
}

function hrtime2ms(h) {
    return h[0] * 1e3 +
           h[1] / 1e6;
}

function hrtime2us(h) {
    return h[0] * 1e6 +
           h[1] / 1e3;
}

/*
 *     S E N D
 * +   M O R E
 * -----------
 *   M O N E Y
 */

function main() {
    // var fs = require('fs');
    // fs.readFileSync('/usr/share/dict/words', 'utf8')

    var search = new ProgSearch(WordProblemState);
    var plan = compileWordProblem('send', 'more', 'money');

    for (var i = 0; i < 15; i++) {
        var start = process.hrtime();
        var results = [];
        search.run(plan, function eachResult(state) {
            results.push(state.result)
            return false;
        });
        var end = process.hrtime();
        console.log(
            'search done in %s (executed %s, expanded %s) found: %j',
            hrtime2us(hrtimeDiff(end, start)), search.executed, search.expanded, results);
    }

}

main();
