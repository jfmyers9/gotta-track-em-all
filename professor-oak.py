import csv

pokemon = []

with open('morepokemon.csv', 'rb') as csvfile:
    spamreader = csv.reader(csvfile, delimiter=',')
    for row in spamreader:
        pokemon.append(row)

inverted = []

for pokeman in pokemon:
   inverted.append([pokeman[0], pokeman[1], str(1.0/float(pokeman[2]))])

for i in inverted:
    print(",".join(i))
