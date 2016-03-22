# File processing #

Basically this boils down to:
- take a big file from http://rdf.dmoz.org/rdf/structure.rdf.u8.gz
- extract all the lines that match the regexp "^\s*<Topic\s+r:id=\"([^"]*)\">\s*$"
- count the average number of slashes in the path(breadcrumbs)
- provide to the user:
  - average number of slashes
  - histogram of the slashes
  - goroutines statistics (how much stuff each core processed)

The most important thing: do everything as quickly as possible/use pararell
processing.

It's not that important here to have proper formatting/packaging, the goal is
to achive highest throughput possible. Also, some extra work was put into not
generating to many objects for the GC - reuse objects as much as possible.

Please check the functions' comments for description of what they do.
