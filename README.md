Batch LCCN Fixer
===

Fixes incorrect LCCNs in chronam / ONI batches:

    go build
    ./batch-lccn-fixer /path/to/some/batch_xxx_yyyyyyy_ver01/ \
      /path/to/fixed/batch_xxx_yyyyyyy_ver01 \
      12345678 \
      sn12345678

This would fix LCCN "12345678" to be "sn12345678" across all files in the
batch, and put it into the "fixed" path.  The fix does the following:

- Rewrites XML by just replacing *all* occurrences of the bad LCCN with the
  good.  This means if the bad LCCN was something like "1", you should not use
  this tool!  It'll simply match too many things it shouldn't.
- Renames any directory paths that are an exact match of the bad LCCN
- Replaces all PDF EXIF data that matches the bad LCCN.  As is the case with
  the XML, if the bad LCCN is something likely to occur in a lot of places,
  this tool must not be used.
