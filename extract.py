import pandas
import re
import sys


def parts(lines):
    buf = []
    for line in lines:
        line = line.rstrip('\r\n')

        sline = line.rstrip()
        if not sline or not sline[0].isspace():
            if buf:
                yield tuple(buf)
                buf = []
            if not sline:
                continue

        buf.append(line)
        continue

    if buf:
        yield tuple(buf)


def extract_prob(lines):
    trace = []
    prob = {
        'trace': trace
    }
    for part in parts(lines):
        match = re.match(r'(\w+)(?:$|:\s*)', part[0])
        if match:
            name = match.group(1)
            rest = part[0][match.end():]
            val = rest
            if len(part) > 1:
                val = (val,) if val else ()
                val += part[1:]
            prob[name] = val
        else:
            trace.append(part)
    return prob


def extract_trace(parts):
    trace_lines = pandas.Series([part[0] for part in parts], name='trace')
    trace = trace_lines.str.extract(r'''
        \[(?P<stack_index>\d+)\]
        \s+
        (?P<op>\w+)
        (?:
            \s+
            (?P<arg>[^\s]+)
        )?
        \s+
        @0x(?P<pi>[0-9a-fA-F]+)
        \s+
        stack=\[
            (?P<stack>.*?)
        \]
    ''', re.VERBOSE)

    trace['stack_index'] = trace['stack_index'].astype('int')
    trace['op'] = trace['op'].astype('category')

    # TODO: y u no base?
    # trace['pi'] = trace['pi'].astype('int', base=16)

    time = extract_timing(parts)
    if time is not None:
        trace = trace.join(time)

    return trace


def extract_timing(parts):
    time_lines = pandas.Series([
        line
        for part in parts
        for line in part
        if line.startswith('  op_time:')
    ])
    if not len(time_lines):
        return None
    time = time_lines.str.extract(r'''
        \s*
        op_time:
        \s+
        clocks=(?P<clocks>\d+)
        \s+
        ns=(?P<ns>\d+)
    ''', re.VERBOSE)
    if not len(time):
        return None
    time['clocks'] = time['clocks'].astype('int')
    time['ns'] = time['ns'].astype('int')
    return time


def dist_table(name, S):
    C = S.value_counts()
    C.index.name = name
    C.name = 'count'
    P = C.astype('float') / len(S)
    P.name = 'pct'
    return pandas.concat([C, P], axis=1)


def extract_program(lines):
    expected = int(re.match(r'\s*- *(\d+)', lines[0]).group(1))
    assert(expected == len(lines[1:]))
    S = pandas.Series(lines[1:])
    D = S.str.extract(r'''
        \s*
        0x(?P<pi>[0-9a-fA-F]+):
        \s+
        (?P<op>\w+)
        (?:
            \s+
            (?P<arg>[^\s]+)
        )?
    ''', re.VERBOSE)
    D['op'] = D['op'].astype('category')
    return D


def extract_prob_timing(prob):
    keys = [
        key
        for key in ('time_setup', 'time_alloc', 'time_run')
        if key in prob
    ]
    if not keys:
        return None

    time_index, time_data = zip(*(
        (key, line)
        for key in keys
        for line in prob[key]
    ))
    if not len(time_data):
        return None

    prob_time = pandas.Series(data=time_data, index=time_index)
    prob_time = prob_time[prob_time.str.endswith(' clocks')]
    if not len(prob_time):
        return None

    prob_time = prob_time.str.rsplit(None, 1).str[0]
    prob_time = prob_time.astype('int')
    prob_time.name = 'clocks'
    return prob_time


prob = extract_prob(sys.stdin)
trace = extract_trace(prob['trace'])
prog = extract_program(prob['program'])

prob_time = extract_prob_timing(prob)
if prob_time is not None:
    print prob_time
    print

print 'plan:'
print '\n'.join(prob['plan'])
print

print 'trace length:', len(trace)
print

print dist_table('stack_index', trace['stack_index'])
print

print dist_table('op', trace['op'])
print

# print dist_table('exitcode', trace[trace['op'] == 'exit']['arg'])
# print

top10_pi = dist_table('pi', trace['pi']).head(10)

# print top10_pi
# print

context = 3
for pi in top10_pi.index:
    cp = top10_pi[top10_pi.index == pi]
    print '%s %i %.1f%%' % (pi, cp['count'][0], cp['pct'][0] * 100)
    i = prog[prog['pi'] == pi].index[0]
    X = prog[i - context:i + context]
    print X
    print

# top10_pi.index

# print 'program:'
# print '\n'.join(prob['program'])
# print

# print trace[trace['op'] == 'push']

if 'clocks' in trace:
    print 'op clocks:'
    print trace.groupby('op')['clocks'].describe().unstack()
