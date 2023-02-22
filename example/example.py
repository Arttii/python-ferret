from pferret import wrapper
import json
compiler = wrapper.Ferret()

with open('example.fql', 'r') as fd:
    fql = fd.read()

params = {
    "take": 10
}
 
res = compiler.execute(fql, params=params)
print(res)
