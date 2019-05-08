potential weaknesses of the AutoPrefixAndValidateSplit validation
* it looks for it prefixed and then not prefixed. it assumes that it
    should be unprefixed if it doesn't exist prefixed
* if the split doesn't exist prefixed or unprefixed, it doesn't care.

so if passed in the merged schema we could maybe feel okay about failing
if the split isn't in the schema in every case (prefixed or otherwise)
and we could allow the force flag to overcome a failed validation here
in the event that the user is certain that this is the right split name
even though it's not in the schema.

try prefixed, try unprefixed, if neither is there, go back to prefixed,
unless the noprefix flag is true

i.e. if noprefix, only check for unprefixed, otherwise, check both and
fallback to prefixed. fail unless force is true.


both false
exists prefixed only   | prefixed
exists unprefixed only | unprefixed
does not exist         | error
both exist             | prefixed

both true
exists prefixed only   | unprefixed
exists unprefixed only | unprefixed
does not exist         | unprefixed
both exist             | unprefixed

force: false, noPrefix: true
exists prefixed only   | error
exists unprefixed only | unprefixed
does not exist         | error
both exist             | unprefixed

force: true, noPrefix: false
exists prefixed only   | prefixed
exists unprefixed only | prefixed
does not exist         | prefixed
both exist             | prefixed
