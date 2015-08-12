'use strict';

var assert = require('assert');
var split2 = require('split2');
var util = require('util');

var StreamSample = require('./stream_sample.js');

var letterBase = 'a'.charCodeAt(0) - 1;

var Operations = {};

Operations.chooseLetter = function chooseLetter(state, op) {
    var start = op.isInitial ? 1 : 0;

    var pend = false;
    var pendDigit = start;

    for (var digit = start; digit < op.base; digit++) {
        if (!state.chosen[digit]) {
            if (pend) {
                var next = (state.alloc()).copyFrom(state);
                next.chosen[pendDigit] = true;
                next.values[op.letter] = pendDigit;
            }
            pend = true;
            pendDigit = digit;
        }
    }

    if (pend) {
        state.chosen[pendDigit] = true;
        state.values[op.letter] = pendDigit;
    } else {
        state.valid = false;
    }
};

Operations.result = function result(state, op) {
    var res = new Array(op.values.length);
    for (var i = 0; i < op.values.length; i++) {
        res[i] = state[op.values[i]];
    }
    state.result = res;
};

Operations.sum = function sum(state, op) {
    var base = op.base;

    var sum = state.carry +
              state.values[op.let1] +
              state.values[op.let2];

    var rem = sum % base;
    state.valid = rem === state.values[op.let3];

    state.carry = Math.floor(sum / base);
};

Operations.checkNoCarry = function checkNoCarry(state, op) {
    state.valid = state.valid &&
                  state.carry === 0;
};

Operations.checkCarry = function checkCarry(state, op) {
    state.valid = state.valid &&
                  state.carry === state.values[op.let3];
};

Operations.toNumber = function toNumber(state, op) {
    var value = 0;
    // for (var i = op.values.length-1; i >= 0; i--)
    for (var i = 0; i < op.word.length; i++) {
        var n = op.word.charCodeAt(i) - letterBase;
        value *= op.base;
        value += state.values[n];
    }
    state[op.store] = value;
};

function lettersFrom(words) {
    var letters = [];
    var seen = {};
    words.forEach(function each(word) {
        for (var i = 0; i < word.length; i++) {
            var n = word.charCodeAt(i) - letterBase;
            if (!seen[n]) {
                seen[n] = true;
                letters.push(n);
            }
        }
    });
    return letters;
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

function WordProblemState() {
    var self = this;

    self.pi = 0;
    self.valid = true;
    self.result = null;
    self.chosen = [];
    self.values = new Uint8Array(26);

    self.carry = 0;
    self.word1 = null;
    self.word2 = null;
    self.word3 = null;
}

WordProblemState.prototype.alloc = function alloc() {
    return new WordProblemState();
};

WordProblemState.prototype.copyFrom = function copyFrom(state) {
    var self = this;

    var i;
    self.pi = state.pi;
    self.valid = true;
    self.result = null;

    self.chosen.length = state.chosen.length;
    for (i = 0; i < state.chosen.length; i++) {
        self.chosen[i] = state.chosen[i];
    }
    for (i = 0; i < self.values.length; i++) {
        self.values[i] = state.values[i];
    }

    self.carry = state.carry;
    self.word1 = state.word1;
    self.word2 = state.word2;
    self.word3 = state.word3;

    return self;
};

function ProgSearch(stateType) {
    var self = this;

    self.stateType = stateType;
    self.freelist = [];
    self.frontier = [];
    self.pushed = 0;
    self.executed = 0;
    self.expanded = 1;

    self.alloc = function alloc() {
        var state;

        if (self.freelist.length) {
            state = self.freelist.shift();
        } else {
            state = new self.stateType();
            state.alloc = self.alloc;
        }

        self.frontier.push(state);
        self.pushed++;

        return state;
    };
}

ProgSearch.prototype.clear = function clear() {
    var self = this;

    if (self.frontier.length) {
        if (self.freelist.length) {
            self.freelist = self.freelist.concat(self.frontier);
        } else {
            self.freelist = self.frontier;
        }
        self.frontier = [];
    }
};

ProgSearch.prototype.run = function run(plan, each) {
    var self = this;

    self.alloc().copyFrom(plan[0].state);
    self.pushed = 0;
    self.executed = 0;
    self.expanded = 1;

    while (self.frontier.length) {
        if (self.expand(plan, each)) {
            break;
        }
    }

    self.clear();
};

ProgSearch.prototype.expand = function expand(plan, each) {
    var self = this;

    var state = self.frontier.shift();
    while (!self.pushed && state.valid && state.pi < plan.length) {
        self.executed++;
        var op = plan[state.pi++];
        op.op(state, op);
    }

    if (!state.valid) {
        self.freelist.push(state);
    } else if (state.result !== null) {
        self.freelist.push(state);
        if (each(state)) {
            return true;
        }
    } else {
        self.frontier.push(state);
        self.pushed++;
    }

    if (self.pushed) {
        self.expanded += self.pushed;
        self.pushed = 0;
        self.heapify();
    }

    return false;
};

ProgSearch.prototype.heapify = function heapify() {
    var self = this;

    for (var i = Math.floor(self.frontier.length / 2 - 1); i >= 0; i--) {
        self.siftdown(i);
    }
};

ProgSearch.prototype.siftup = function siftup(i) {
    var self = this;

    while (i > 0) {
        var j = Math.floor((i - 1) / 2);
        var child = self.frontier[i];
        var par = self.frontier[j];
        if (child.pi > par.pi) {
            self.frontier[i] = par
            self.frontier[j] = child;
            i = j;
        }
    }
};

ProgSearch.prototype.siftdown = function siftdown(i) {
    var self = this;

    while (true) {
        var par = self.frontier[i];

        // left
        var j = (2 * i) + 1;
        if (j >= self.frontier.length) {
            return;
        }

        // maybe right
        var child = self.frontier[j];
        if (++j >= self.frontier.length ||
            self.frontier[j].pi <= child.pi) {
            j--;
        } else {
            child = self.frontier[j];
        }

        if (child.pi <= par.pi) {
            return;
        }

        self.frontier[i] = child;
        self.frontier[j] = par;

        i = j;
    }
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

function WordProblem() {
    this.word1 = '';
    this.word2 = '';
    this.word3 = '';
    this.result = null;
    this.time = [0, 0];
    this.executed = 0;
    this.expanded = 0;
    this.plan = null;
    this.base = 0;
};

WordProblem.prototype.reset = function reset() {
    this.result = null;
    this.time = [0, 0];
    this.executed = 0;
    this.expanded = 0;
};

WordProblem.prototype.compile = function compile() {
    var self = this;

    if (self.plan) {
        self.reset();
        self.plan = null;
    }

    var lenDiff = self.word3.length - self.word2.length;
    if (self.word1.length > 8) {
        return;
    }

    if (self.word1.length !== self.word2.length) {
        return;
    }

    if (lenDiff !== 0 && lenDiff !== 1) {
        return;
    }

    self.word1 = self.word1.toLowerCase();
    self.word2 = self.word2.toLowerCase();
    self.word3 = self.word3.toLowerCase();

    var letters = lettersFrom([self.word1, self.word2, self.word3]);
    self.base = baseFor(letters.length);

    var plan = [];

    var initialState = new WordProblemState();
    initialState.chosen.length = self.base;
    for (var i = 0; i < self.base; i++) {
        initialState.chosen[i] = false;
    }
    initialState.pi = 1;

    plan.push({
        state: initialState
    });

    var seen = {};
    for (var i = 1; i <= self.word1.length; i++) {
        plan.push({
            op: Operations.sum,
            base: self.base,
            let1: addLetter(self.word1, self.word1.length - i),
            let2: addLetter(self.word2, self.word2.length - i),
            let3: addLetter(self.word3, self.word3.length - i)
        });
    }

    if (lenDiff) {
        plan.push({
            op: Operations.checkCarry,
            let3: addLetter(self.word3, 0)
        });
    } else {
        plan.push({
            op: Operations.checkNoCarry
        });
    }

    plan.push({
        op: Operations.toNumber,
        store: 'word1',
        word: self.word1,
        base: self.base
    });

    plan.push({
        op: Operations.toNumber,
        store: 'word2',
        word: self.word2,
        base: self.base
    });

    plan.push({
        op: Operations.toNumber,
        store: 'word3',
        word: self.word3,
        base: self.base
    });

    plan.push({
        op: Operations.result,
        values: ['word1', 'word2', 'word3']
    });

    self.plan = plan;

    function addLetter(word, i) {
        var c = word.charCodeAt(i);
        var n = c - letterBase;
        if (!seen[n]) {
            seen[n] = true;
            plan.push({
                op: Operations.chooseLetter,
                letter: n,
                base: self.base,
                isInitial: i === 0 || c === word.charCodeAt(0)
            });
        }
        return n;
    }
};

WordProblem.prototype.run = function run(search) {
    var start = process.hrtime();
    this.compile();
    if (this.plan) {
        this.runPlan(search);
    }
    this.time = hrtimeDiff(process.hrtime(), start);
};

WordProblem.prototype.runPlan = function runPlan(search) {
    var self = this;

    search.run(self.plan, eachResult);
    self.executed = search.executed;
    self.expanded = search.expanded;

    function eachResult(state) {
        self.result = state.result;
        return true;
    }
};

function printSol(sol) {
    console.log('%s(%s, %s, %s) in %sus result: %s ',
                sol.plan ? 'solved' : 'skipped',
                sol.word1, sol.word2, sol.word3,
                hrtime2us(sol.time),
                sol.result);
}

function find(words, each) {
    var search = new ProgSearch(WordProblemState);
    var prob = new WordProblem();
    for (var i = 0; i < words.length; i++) {
        prob.word1 = words[i];
        for (var j = i + 1; j < words.length; j++) {
            prob.word2 = words[j];
            for (var k = 0; k < words.length; k++) {
                prob.word3 = words[k];
                prob.run(search);
                each(prob);
            }
        }
    }
}

function searchStream(stream) {
    var lines = [];
    stream
        .pipe(split2())
        .on('data', function each(line) {
            lines.push(line);
        })
        .on('end', function readDone() {
            find(lines, each);
        });

    var n = 0;
    var attempted = 0;
    var skipped = 0;
    var found = 0;
    function each(sol) {
        // prob -> sol
        if (sol.plan) {
            attempted++;
            if (sol.result) {
                found++;
                printSol(sol);
            }
        } else {
            skipped++;
        }
        if (++n % 1000 === 0) {
            console.log(
                'attempted %s skipped %s found %s',
                attempted, skipped, found);
        }
    }
}

function test() {
    var search = new ProgSearch(WordProblemState);
    var prob = new WordProblem();
    prob.word1 = 'send';
    prob.word2 = 'more';
    prob.word3 = 'money';
    for (var i = 0; i < 5; i++) {
        prob.run(search);
        printSol(prob);
    }
}

function testFind(stream, n, seed) {
    if (typeof n !== 'number' || isNaN(n) ||
        typeof seed !== 'number' || isNaN(seed)
    ) {
        console.error('usage: testFind N SEED');
        process.exit(1);
    }

    StreamSample(stream.pipe(split2()), n, seed, onSample);

    function onSample(err, sample) {
        var words = new Array(sample.length);
        for (var i = 0; i < sample.length; i++) {
            words[i] = sample[i].item;
        }

        find(words, printSol);
    }
}

function main(argv) {
    if (argv[1] === 'test') {
        test();
    } else if (argv[1] === 'testFind') {
        testFind(process.stdin, parseInt(argv[2]), parseInt(argv[3]));
    } else {
        searchStream(process.stdin);
    }
}

main(process.argv.slice(1));
