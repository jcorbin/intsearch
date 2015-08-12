'use strict';

var assert = require('assert');
var split2 = require('split2');
var util = require('util');

var StreamSample = require('./stream_sample.js');

var letterBase = 'a'.charCodeAt(0) - 1;

function poolize(Cons) {
    Cons.pool = [];

    Cons.alloc = function alloc() {
        var self;
        if (Cons.pool.length) {
            self = Cons.pool.shift();
        } else {
            self = new Cons();
        }
        return self;
    };

    Cons.free = function free(self) {
        Cons.pool.push(self);
    };

    Cons.prototype.free = function free() {
        Cons.pool.push(this);
    };

}

function InitialStateOperaion() {
    this.state = null;
}

poolize(InitialStateOperaion);

InitialStateOperaion.prototype.init = function init(state) {
    if (state !== this.state) {
        this.state = state;
    }
    return this;
};

InitialStateOperaion.prototype.run = function initialState(state) {
    var pi = state.pi;
    state.copyFrom(this.state);
    state.pi = pi;
};

function ChooseLetterOperation() {
    this.letter = '';
    this.base = 2;
    this.start = 0;
}

poolize(ChooseLetterOperation);

ChooseLetterOperation.prototype.init = function init(letter, base, isInitial) {
    this.letter = letter;
    this.base = base;
    this.start = isInitial ? 1 : 0;
    return this;
};

ChooseLetterOperation.prototype.run = function chooseLetter(state) {
    var pend = false;
    var pendDigit = this.start;

    for (var digit = this.start; digit < this.base; digit++) {
        if (!state.chosen[digit]) {
            if (pend) {
                var next = (state.alloc()).copyFrom(state);
                next.chosen[pendDigit] = true;
                next.values[this.letter] = pendDigit;
            }
            pend = true;
            pendDigit = digit;
        }
    }

    if (pend) {
        state.chosen[pendDigit] = true;
        state.values[this.letter] = pendDigit;
    } else {
        state.valid = false;
    }
};

function ResultOperation() {
    this.values = null;
}

ResultOperation.prototype.init = function init(values) {
    this.values = values;
    return this;
};

poolize(ResultOperation);

ResultOperation.prototype.run = function result(state) {
    var res = new Array(this.values.length);
    for (var i = 0; i < this.values.length; i++) {
        res[i] = state[this.values[i]];
    }
    state.result = res;
};

function SumOperation() {
    this.let1 = '';
    this.let2 = '';
    this.let3 = '';
    this.base = 2;
}

poolize(SumOperation);

SumOperation.prototype.init = function init(let1, let2, let3, base) {
    this.let1 = let1;
    this.let2 = let2;
    this.let3 = let3;
    this.base = base;
    return this;
};

SumOperation.prototype.run = function sum(state) {
    var sum = state.carry +
              state.values[this.let1] +
              state.values[this.let2];

    var rem = sum % this.base;
    state.valid = rem === state.values[this.let3];

    state.carry = Math.floor(sum / this.base);
};

function CheckCarryOperation() {
    this.letter = '';
}

poolize(CheckCarryOperation);

CheckCarryOperation.prototype.init = function init(letter) {
    this.letter = letter;
    this.run = this.letter ? this.runWithLetter : this.runNoLetter;
    return this;
};

CheckCarryOperation.prototype.runWithLetter = function checkCarry(state) {
    state.valid = state.valid &&
                  state.carry === state.values[this.letter];
};

CheckCarryOperation.prototype.runNoLetter = function checkNoCarry(state) {
    state.valid = state.valid &&
                  state.carry === 0;
};

function ToNumberOperation() {
    this.word = '';
    this.base = 0;
    this.store = '';
}

poolize(ToNumberOperation);

ToNumberOperation.prototype.init = function init(word, base, store) {
    this.word = word;
    this.base = base;
    this.store = store;
    return this;
};

ToNumberOperation.prototype.run = function toNumber(state) {
    var value = 0;
    for (var i = 0; i < this.word.length; i++) {
        var n = this.word.charCodeAt(i) - letterBase;
        value *= this.base;
        value += state.values[n];
    }
    state[this.store] = value;
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

poolize(WordProblemState);

WordProblemState.prototype.alloc = function alloc() {
    return new WordProblemState();
};

WordProblemState.prototype.reset = function reset() {
    return this.init(this.chosen.length);
};

WordProblemState.prototype.init = function init(base) {
    var i;

    this.pi = 0;
    this.valid = true;
    this.result = null;

    this.chosen.length = base;
    for (i = 0; i < base; i++) {
        this.chosen[i] = false;
    }

    for (i = 0; i < this.values.length; i++) {
        this.values[i] = 0;
    }

    this.carry = 0;
    this.word1 = null;
    this.word2 = null;
    this.word3 = null;

    return this;
};

WordProblemState.prototype.copyFrom = function copyFrom(state) {
    var i;

    this.pi = state.pi;
    this.valid = true;
    this.result = null;

    this.chosen.length = state.chosen.length;
    for (i = 0; i < state.chosen.length; i++) {
        this.chosen[i] = state.chosen[i];
    }
    for (i = 0; i < this.values.length; i++) {
        this.values[i] = state.values[i];
    }

    this.carry = state.carry;
    this.word1 = state.word1;
    this.word2 = state.word2;
    this.word3 = state.word3;

    return this;
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

    var state = self.alloc().reset();
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
        op.run(state);
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
        self.freePlan();
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

    var initialState = (new WordProblemState()).init(self.base);
    plan.push(InitialStateOperaion.alloc().init(initialState));

    var seen = {};
    for (var i = 1; i <= self.word1.length; i++) {
        var let1 = addLetter(self.word1, self.word1.length - i);
        var let2 = addLetter(self.word2, self.word2.length - i);
        var let3 = addLetter(self.word3, self.word3.length - i);
        plan.push(SumOperation.alloc().init(let1, let2, let3, self.base));
    }
    var lastLetter = lenDiff ? addLetter(self.word3, 0) : '';
    plan.push(CheckCarryOperation.alloc().init(lastLetter));

    plan.push(ToNumberOperation.alloc().init(self.word1, self.base, 'word1'));
    plan.push(ToNumberOperation.alloc().init(self.word2, self.base, 'word2'));
    plan.push(ToNumberOperation.alloc().init(self.word3, self.base, 'word3'));
    plan.push(ResultOperation.alloc().init(['word1', 'word2', 'word3']));

    self.plan = plan;

    function addLetter(word, i) {
        var c = word.charCodeAt(i);
        var n = c - letterBase;
        if (!seen[n]) {
            seen[n] = true;
            var isInitial = i === 0 || c === word.charCodeAt(0);
            plan.push(ChooseLetterOperation.alloc().init(n, self.base, isInitial));
        }
        return n;
    }
};

WordProblem.prototype.freePlan = function freePlan() {
    if (this.plan) {
        for (var i = 0; i < this.plan.length; i++) {
            this.plan[i].free();
        }
        this.plan = null;
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
