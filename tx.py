import simplejson as json


def printR(res):
    print(json.dumps(res))


tx = {}
with open("./tx", "rb") as _tx:
    tx = json.loads(_tx.read())

print(tx)
print()
printR(list(tx["result"].keys()))
print()
printR(tx["result"]["transaction"])
print()
printR(tx["result"]["meta"]["postTokenBalances"])

