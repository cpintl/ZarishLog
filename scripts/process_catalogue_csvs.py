#!/usr/bin/env python3
"""
ZarishLog — Product Catalogue Generator

Takes one or more raw source CSVs (e.g. scraped national drug-database dumps)
and normalizes them into the checked-in `master_product_catalogue.csv` schema
used by `make db-seed` and the configuration validator.

SKU scheme: ``{item_type}-{slugified name}-{seq:04d}`` (e.g. ``DRUG-acemetacin-0001``).

Usage::

    python3 scripts/process_catalogue_csvs.py path/to/source.csv
    python3 scripts/process_catalogue_csvs.py a.csv b.csv --out config/metadata/master_product_catalogue.csv
    python3 scripts/process_catalogue_csvs.py --rebuild-from catalogue.csv   # re-slug existing catalogue

Input columns accepted (any subset; the rest default sensibly):
    name, category_name, item_type, uom_abbreviation, description, strength,
    dosage_form, manufacturers, source, is_asset

This script is idempotent: re-running overwrites the output file cleanly.
"""

from __future__ import annotations

import argparse
import csv
import os
import re
import sys
from collections import Counter
from pathlib import Path

HEADER = [
    "sku",
    "name",
    "category_name",
    "item_type",
    "description",
    "strength",
    "dosage_form",
    "uom_abbreviation",
    "is_batch_tracked",
    "is_expiry_tracked",
    "is_cold_chain",
    "is_hazardous",
    "is_essential",
    "is_controlled",
    "is_asset",
    "replenishment_type",
    "valuation_method",
    "manufacturers",
    "source",
]

ASSET_CATEGORIES = {"equipment", "assets", "non-medical", "furniture"}
CONTROLLED_NAMES = ("morphine", "diazepam", "pethidine", "fentanyl", "ketamine", "tramadol", "codeine")

COLD_CHAIN_NAMES = ("vaccine", "insulin", "cold", "refrigerat", "cryo", "tetanus", "diphtheria", "bcg")


def slugify(name: str) -> str:
    name = (name or "").lower().strip()
    name = re.sub(r"[^a-z0-9]+", "-", name)
    return name.strip("-")[:60]


def sniff_bool(row: dict, key: str) -> str:
    val = (row.get(key) or "").strip().lower()
    return "TRUE" if val in {"1", "true", "yes", "y", "t"} else "FALSE"


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("sources", nargs="+", help="Input CSV file(s) to normalize")
    parser.add_argument(
        "--out",
        default="config/metadata/master_product_catalogue.csv",
        help="Output master catalogue path",
    )
    parser.add_argument("--suppress-header-consistency-check", action="store_true")
    args = parser.parse_args()

    rows: list[dict] = []
    for path in args.sources:
        with open(path, newline="", encoding="utf-8-sig") as fh:
            reader = csv.DictReader(fh)
            if reader.fieldnames is None:
                print(f"skipping empty file: {path}")
                continue
            for raw in reader:
                name = (raw.get("name") or "").strip()
                if not name:
                    continue
                rows.append(raw)

    if not rows:
        print("error: no rows read from input files")
        return 1

    seen_names: Counter[str] = Counter()
    out_rows = []
    for raw in rows:
        name = (raw.get("name") or "").strip()
        item_type = (raw.get("item_type") or "drug").strip().lower()
        category = (raw.get("category_name") or "Uncategorized").strip()
        seen_names[name] += 1
        key = slugify(name)
        if seen_names[name] > 1:
            key = f"{key}-{seen_names[name]}"
        seq = len(out_rows) + 1

        desc = raw.get("description") or name
        strength = (raw.get("strength") or "").strip()
        dosage = (raw.get("dosage_form") or "").strip()
        if not desc and (strength or dosage):
            desc = f"{name} | {f'({strength}) ' if strength else ''}Dosage: {dosage}"

        lower_name = name.lower()
        is_asset = "TRUE" if category.lower().split()[0] in ASSET_CATEGORIES or raw.get("is_asset") == "1" else "FALSE"
        is_controlled = "TRUE" if any(w in lower_name for w in CONTROLLED_NAMES) else "FALSE"
        is_cold = "TRUE" if any(w in lower_name for w in COLD_CHAIN_NAMES) else "FALSE"

        out_rows.append(
            {
                "sku": f"{item_type}-{key}-{seq:04d}",
                "name": name,
                "category_name": category,
                "item_type": item_type,
                "description": desc,
                "strength": strength,
                "dosage_form": dosage,
                "uom_abbreviation": (raw.get("uom_abbreviation") or "EA").strip(),
                "is_batch_tracked": sniff_bool(raw, "is_batch_tracked") or "TRUE",
                "is_expiry_tracked": sniff_bool(raw, "is_expiry_tracked") or "TRUE",
                "is_cold_chain": is_cold,
                "is_hazardous": sniff_bool(raw, "is_hazardous"),
                "is_essential": sniff_bool(raw, "is_essential"),
                "is_controlled": is_controlled,
                "is_asset": is_asset,
                "replenishment_type": (raw.get("replenishment_type") or "min_max").strip(),
                "valuation_method": (raw.get("valuation_method") or "fifo").strip(),
                "manufacturers": (raw.get("manufacturers") or "").strip(),
                "source": (raw.get("source") or os.path.basename(path)).strip(),
            }
        )

    out_path = Path(args.out)
    out_path.parent.mkdir(parents=True, exist_ok=True)
    with open(out_path, "w", newline="", encoding="utf-8") as fh:
        writer = csv.DictWriter(fh, fieldnames=HEADER, quoting=csv.QUOTE_ALL)
        writer.writeheader()
        writer.writerows(out_rows)

    by_type = Counter(r["item_type"] for r in out_rows)
    print(f"wrote {len(out_rows)} products -> {out_path}")
    for item_type, count in by_type.most_common():
        print(f"  {item_type}: {count}")
    return 0


if __name__ == "__main__":
    sys.exit(main())