#!/usr/bin/env python3
"""
ZarishLog — Process raw CSV files into Master Product Catalogue
Inputs: 5 CSV files from /home/codeandbrain/Documents/
Outputs: config/metadata/master_product_catalogue.csv + seed SQL
"""

import csv
import hashlib
import json
import os
import re
import sys
from collections import defaultdict
from typing import Optional

SCRIPT_DIR = os.path.dirname(os.path.abspath(__file__))
DOCS_DIR = os.path.join(SCRIPT_DIR, "..", "input")
CONFIG_DIR = os.path.realpath(os.path.join(SCRIPT_DIR, "..", "config", "metadata"))

# ─── Drug Therapeutic Classification ──────────────────────────────────────

# Explicit drug name → category mapping (for drugs regex can't easily catch)
EXPLICIT_DRUG_MAP = {
    # Oncology (targeted therapies, immunomodulators, chemotherapies)
    "afatinib": "Oncology", "alectinib": "Oncology", "alpelisib": "Oncology",
    "axitinib": "Oncology", "avelumab": "Oncology", "avapritinib": "Oncology",
    "azacitidine": "Oncology", "bevacizumab": "Oncology", "binimetinib": "Oncology",
    "bortezomib": "Oncology", "bosutinib": "Oncology", "busulfan": "Oncology",
    "cabazitaxel": "Oncology", "cabozantinib": "Oncology", "carmustine": "Oncology",
    "ceritinib": "Oncology", "cetuximab": "Oncology", "chlorambucil": "Oncology",
    "cobimetinib": "Oncology", "crizotinib": "Oncology", "cyclophosphamide": "Oncology",
    "cytarabine": "Oncology", "dabrafenib": "Oncology", "dacarbazine": "Oncology",
    "dactinomycin": "Oncology", "dasatinib": "Oncology", "daunorubicin": "Oncology",
    "decitabine": "Oncology", "docetaxel": "Oncology", "doxorubicin": "Oncology",
    "durvalumab": "Oncology", "encorafenib": "Oncology", "enzalutamide": "Oncology",
    "epirubicin": "Oncology", "eribulin": "Oncology", "erlotinib": "Oncology",
    "etoposide": "Oncology", "everolimus": "Oncology", "exemestane": "Oncology",
    "fludarabine": "Oncology", "fluorouracil": "Oncology", "fulvestrant": "Oncology",
    "gefitinib": "Oncology", "gemcitabine": "Oncology", "goserelin": "Oncology",
    "hydroxyurea": "Oncology", "hydroxycarbamide": "Oncology", "ibrutinib": "Oncology",
    "idarubicin": "Oncology", "ifosfamide": "Oncology", "imatinib": "Oncology",
    "irinotecan": "Oncology", "ixazomib": "Oncology", "lapatinib": "Oncology",
    "lenalidomide": "Oncology", "lenvatinib": "Oncology", "leuprolide": "Oncology",
    "lomustine": "Oncology", "methotrexate": "Oncology",
    "mitomycin": "Oncology", "mitoxantrone": "Oncology",
    "navelbine": "Oncology", "neratinib": "Oncology",
    "nilotinib": "Oncology", "nintedanib": "Oncology", "niraparib": "Oncology",
    "olaparib": "Oncology", "osimertinib": "Oncology", "oxaliplatin": "Oncology",
    "paclitaxel": "Oncology", "palbociclib": "Oncology", "panitumumab": "Oncology",
    "panobinostat": "Oncology", "pazopanib": "Oncology", "pembrolizumab": "Oncology",
    "pemetrexed": "Oncology", "pertuzumab": "Oncology", "pomalidomide": "Oncology",
    "ponatinib": "Oncology", "regorafenib": "Oncology", "ribociclib": "Oncology",
    "rituximab": "Oncology", "rucaparib": "Oncology", "ruxolitinib": "Oncology",
    "sorafenib": "Oncology", "sunitinib": "Oncology", "talazoparib": "Oncology",
    "tamoxifen": "Oncology", "tegafur": "Oncology", "temozolomide": "Oncology",
    "temsirolimus": "Oncology", "thalidomide": "Oncology",
    "topotecan": "Oncology", "trametinib": "Oncology", "trastuzumab": "Oncology",
    "vandetanib": "Oncology", "vemurafenib": "Oncology", "vinblastine": "Oncology",
    "vincristine": "Oncology", "vindesine": "Oncology", "vinorelbine": "Oncology",
    "vismodegib": "Oncology", "vorinostat": "Oncology",
    "asciminib": "Oncology",
    # Immunosuppressants
    "azathioprine": "Immunosuppressants", "mycophenolate": "Immunosuppressants",
    "cyclosporine": "Immunosuppressants", "ciclosporin": "Immunosuppressants",
    "tacrolimus": "Immunosuppressants", "sirolimus": "Immunosuppressants",
    "everolimus": "Immunosuppressants", "basiliximab": "Immunosuppressants",
    "baricitinib": "Immunosuppressants", "tofacitinib": "Immunosuppressants",
    "upadacitinib": "Immunosuppressants",
    # Musculoskeletal / Bone
    "alendronic": "Musculoskeletal", "alendronate": "Musculoskeletal",
    "risedronate": "Musculoskeletal", "zoledronic": "Musculoskeletal",
    "ibandronic": "Musculoskeletal", "calcitonin": "Musculoskeletal",
    "glucosamine": "Musculoskeletal", "chondroitin": "Musculoskeletal",
    "baclofen": "Musculoskeletal", "tolperisone": "Musculoskeletal",
    "alfacalcidol": "Musculoskeletal", "calcitriol": "Musculoskeletal",
    # Pulmonary hypertension
    "ambrisentan": "Cardiovascular", "bosentan": "Cardiovascular",
    "sildenafil": "Cardiovascular", "tadalafil": "Cardiovascular",
    "iloprost": "Cardiovascular", "riociguat": "Cardiovascular",
    "selexipag": "Cardiovascular", "treprostinil": "Cardiovascular",
    # Immunoglobulins / Blood products
    "albumin": "IV Fluids & Electrolytes", "alteplase": "Cardiovascular",
    "tenecteplase": "Cardiovascular", "streptokinase": "Cardiovascular",
    "urokinase": "Cardiovascular", "factor viii": "Hemostatics",
    "factor ix": "Hemostatics", "enoxaparin": "Anticoagulants",
    "dalteparin": "Anticoagulants", "fondaparinux": "Anticoagulants",
    "argatroban": "Anticoagulants", "bivalirudin": "Anticoagulants",
    # Antidiabetic incretins / newer agents
    "alogliptin": "Antidiabetics", "linagliptin": "Antidiabetics",
    "dulaglutide": "Antidiabetics", "exenatide": "Antidiabetics",
    "liraglutide": "Antidiabetics", "semaglutide": "Antidiabetics",
    "lixisenatide": "Antidiabetics",
    # Psychotropics (atypical antipsychotics, mood stabilizers)
    "aripiprazole": "CNS & Psychotropics", "asenapine": "CNS & Psychotropics",
    "lurasidone": "CNS & Psychotropics", "paliperidone": "CNS & Psychotropics",
    "ziprasidone": "CNS & Psychotropics", "cariprazine": "CNS & Psychotropics",
    "brexpiprazole": "CNS & Psychotropics", "agomelatine": "CNS & Psychotropics",
    "armodafinil": "CNS & Psychotropics", "modafinil": "CNS & Psychotropics",
    "lithium": "CNS & Psychotropics", "lamotrigine": "CNS & Psychotropics",
    "oxcarbazepine": "CNS & Psychotropics", "vigabatrin": "CNS & Psychotropics",
    # Urology
    "tamsulosin": "Urologicals", "alfuzosin": "Urologicals",
    "dutasteride": "Urologicals", "finasteride": "Urologicals",
    "solifenacin": "Urologicals", "tolterodine": "Urologicals",
    "oxybutynin": "Urologicals", "mirabegron": "Urologicals",
    "sildenafil": "Urologicals", "tadalafil": "Urologicals",
    "vardenafil": "Urologicals", "avanafil": "Urologicals",
    # Diagnostic agents
    "iohexol": "Diagnostic Contrast Media", "iopamidol": "Diagnostic Contrast Media",
    "ioversol": "Diagnostic Contrast Media", "gadobutrol": "Diagnostic Contrast Media",
    "gadoteric": "Diagnostic Contrast Media", "barium sulfate": "Diagnostic Contrast Media",
    # Antifungals
    "anidulafungin": "Systemic Antifungals", "caspofungin": "Systemic Antifungals",
    "micafungin": "Systemic Antifungals", "amorolfine": "Dermatologicals",
    # Various specific drugs
    "acetazolamide": "Cardiovascular",
    "activated charcoal": "Antidotes & Antivenoms",
    "adrenaline": "Anesthetics",
    "alprostadil": "Cardiovascular",
    "aminocaproic acid": "Hemostatics",
    "baloxavir marboxil": "Antivirals",
    "bambuterol": "Respiratory",
    "benzonatate": "Respiratory",
    "bimatoprost": "Ophthalmologicals",
    "brimonidine": "Ophthalmologicals",
    "brinzolamide": "Ophthalmologicals",
    "bumetanide": "Cardiovascular",
    "carglumic acid": "Gastrointestinal",
    "clopidogrel": "Cardiovascular",
    "colchicine": "Musculoskeletal",
    "dabigatran": "Anticoagulants",
    "dantrolene": "Musculoskeletal",
    "deferasirox": "Antidotes & Antivenoms",
    "deferiprone": "Antidotes & Antivenoms",
    "desmopressin": "Hormones & Corticosteroids",
    "dexmedetomidine": "Anesthetics",
    "dinoprostone": "Obstetrics & Gynecology",
    "dornase alfa": "Respiratory",
    "edoxaban": "Anticoagulants",
    "eletriptan": "CNS & Psychotropics",
    "eltrombopag": "Hemostatics",
    "emedastine": "Ophthalmologicals",
    "ephedrine": "Respiratory",
    "epinephrine": "Anesthetics",
    "eplerenone": "Cardiovascular",
    "ergometrine": "Obstetrics & Gynecology",
    "ergotamine": "CNS & Psychotropics",
    "eslicarbazepine": "CNS & Psychotropics",
    "esomeprazole": "Gastrointestinal",
    "etamsylate": "Hemostatics",
    "etravirine": "Antiretrovirals",
    "ezetimibe": "Cardiovascular",
    "febuxostat": "Musculoskeletal",
    "fidaxomicin": "Antibiotics",
    "flavoxate": "Urologicals",
    "flumazenil": "Antidotes & Antivenoms",
    "flurbiprofen": "Analgesics & Anti-inflammatory",
    "fomepizole": "Antidotes & Antivenoms",
    "frovatriptan": "CNS & Psychotropics",
    "gan ciclovir": "Antivirals",
    "glyceryl trinitrate": "Cardiovascular",
    "glycopyrrolate": "Gastrointestinal",
    "granisetron": "Gastrointestinal",
    "hyaluronic acid": "Musculoskeletal",
    "hyoscyamine": "Gastrointestinal",
    "icatibant": "Antidotes & Antivenoms",
    "idebenone": "CNS & Psychotropics",
    "idursulfase": "Metabolic Disorders",
    "iloperidone": "CNS & Psychotropics",
    "indacaterol": "Respiratory",
    "isosorbide": "Cardiovascular",
    "isosorbide dinitrate": "Cardiovascular",
    "isosorbide mononitrate": "Cardiovascular",
    "isoxsuprine": "Cardiovascular",
    "ivabradine": "Cardiovascular",
    "ketorolac": "Analgesics & Anti-inflammatory",
    "lacosamide": "CNS & Psychotropics",
    "lanreotide": "Oncology",
    "latanoprost": "Ophthalmologicals",
    "levalbuterol": "Respiratory",
    "levobunolol": "Ophthalmologicals",
    "levocarnitine": "Metabolic Disorders",
    "levocetirizine": "Antihistamines & Allergy",
    "levodopa": "CNS & Psychotropics",
    "levofloxacin": "Antibiotics",
    "levosimendan": "Cardiovascular",
    "lidocaine": "Anesthetics",
    "lignocaine": "Anesthetics",
    "linezolid": "Antibiotics",
    "lomitapide": "Cardiovascular",
    "loperamide": "Gastrointestinal",
    "lopinavir": "Antiretrovirals",
    "lorcaserin": "CNS & Psychotropics",
    "lubiprostone": "Gastrointestinal",
    "mebeverine": "Gastrointestinal",
    "mecobalamin": "Vitamins & Minerals",
    "melatonin": "CNS & Psychotropics",
    "meropenem": "Antibiotics",
    "mesalamine": "Gastrointestinal",
    "mesalazine": "Gastrointestinal",
    "metaraminol": "Cardiovascular",
    "methylprednisolone": "Hormones & Corticosteroids",
    "metoclopramide": "Gastrointestinal",
    "metoprolol": "Cardiovascular",
    "metronidazole": "Antibiotics",
    "mexiletine": "Cardiovascular",
    "midazolam": "Anesthetics",
    "midodrine": "Cardiovascular",
    "milrinone": "Cardiovascular",
    "mitotane": "Oncology",
    "miVacurium": "Anesthetics",
    "mometasone": "Respiratory",
    "moxifloxacin": "Antibiotics",
    "mycophenolic acid": "Immunosuppressants",
    "nalbuphine": "Analgesics & Anti-inflammatory",
    "nalidixic acid": "Antibiotics",
    "naloxone": "Antidotes & Antivenoms",
    "naphazoline": "Ophthalmologicals",
    "naproxen": "Analgesics & Anti-inflammatory",
    "naratriptan": "CNS & Psychotropics",
    "nateglinide": "Antidiabetics",
    "neomycin": "Antibiotics",
    "neostigmine": "Anesthetics",
    "nepafenac": "Ophthalmologicals",
    "nesiritide": "Cardiovascular",
    "nevirapine": "Antiretrovirals",
    "niacin": "Vitamins & Minerals",
    "nicorandil": "Cardiovascular",
    "nifedipine": "Cardiovascular",
    "nimesulide": "Analgesics & Anti-inflammatory",
    "nimodipine": "Cardiovascular",
    "nitazoxanide": "Antibiotics",
    "nitrofurantoin": "Antibiotics",
    "nitroglycerin": "Cardiovascular",
    "nizatidine": "Gastrointestinal",
    "norepinephrine": "Anesthetics",
    "norethisterone": "Hormones & Corticosteroids",
    "norgestrel": "Contraceptives",
    "nortriptyline": "CNS & Psychotropics",
    "nystatin": "Systemic Antifungals",
    "octreotide": "Gastrointestinal",
    "ofloxacin": "Antibiotics",
    "olanzapine": "CNS & Psychotropics",
    "olmesartan": "Cardiovascular",
    "olodaterol": "Respiratory",
    "olopatadine": "Ophthalmologicals",
    "omalizumab": "Respiratory",
    "omeprazole": "Gastrointestinal",
    "ondansetron": "Gastrointestinal",
    "orciprenaline": "Respiratory",
    "orlistat": "Gastrointestinal",
    "oseltamivir": "Antivirals",
    "oxaliplatin": "Oncology",
    "oxcarbazepine": "CNS & Psychotropics",
    "oxprenolol": "Cardiovascular",
    "oxybutynin": "Urologicals",
    "oxycodone": "Analgesics & Anti-inflammatory",
    "oxymetazoline": "ENT Preparations",
    "oxytocin": "Obstetrics & Gynecology",
    "palmidrol": "Analgesics & Anti-inflammatory",
    "pancuronium": "Anesthetics",
    "pancreatin": "Gastrointestinal",
    "pantoprazole": "Gastrointestinal",
    "paracetamol": "Analgesics & Anti-inflammatory",
    "parecoxib": "Analgesics & Anti-inflammatory",
    "paricalcitol": "Renal & Dialysis",
    "paroxetine": "CNS & Psychotropics",
    "pemetrexed": "Oncology",
    "pencilamine": "Antidotes & Antivenoms",
    "penicillin": "Antibiotics",
    "pentamidine": "Antibiotics",
    "pentazocine": "Analgesics & Anti-inflammatory",
    "pentobarbital": "CNS & Psychotropics",
    "pentoxifylline": "Cardiovascular",
    "perindopril": "Cardiovascular",
    "perphenazine": "CNS & Psychotropics",
    "pethidine": "Analgesics & Anti-inflammatory",
    "phenelzine": "CNS & Psychotropics",
    "pheniramine": "Antihistamines & Allergy",
    "phenobarbital": "CNS & Psychotropics",
    "phenoxybenzamine": "Cardiovascular",
    "phensuximide": "CNS & Psychotropics",
    "phentermine": "CNS & Psychotropics",
    "phentolamine": "Cardiovascular",
    "phenylbutyrate": "Metabolic Disorders",
    "phenylephrine": "Cardiovascular",
    "phenytoin": "CNS & Psychotropics",
    "phosphomycin": "Antibiotics",
    "phytonadione": "Vitamins & Minerals",
    "pilocarpine": "Ophthalmologicals",
    "pimavanserin": "CNS & Psychotropics",
    "pimozide": "CNS & Psychotropics",
    "pindolol": "Cardiovascular",
    "pioglitazone": "Antidiabetics",
    "piperacillin": "Antibiotics",
    "pipothiazine": "CNS & Psychotropics",
    "pirbuterol": "Respiratory",
    "piretanide": "Cardiovascular",
    "piribedil": "CNS & Psychotropics",
    "piroxicam": "Analgesics & Anti-inflammatory",
    "pitavastatin": "Cardiovascular",
    "pneumococcal": "Vaccines & Immunoglobulins",
    "podophyllotoxin": "Dermatologicals",
    "polymyxin": "Antibiotics",
    "polysaccharide iron": "Vitamins & Minerals",
    "ponatinib": "Oncology",
    "posaconazole": "Systemic Antifungals",
    "potassium chloride": "IV Fluids & Electrolytes",
    "potassium citrate": "Urologicals",
    "povidone iodine": "Dermatologicals",
    "pramipexole": "CNS & Psychotropics",
    "pranlukast": "Respiratory",
    "prasugrel": "Cardiovascular",
    "pravastatin": "Cardiovascular",
    "praziquantel": "Antihelminthics",
    "prazosin": "Cardiovascular",
    "prednisolone": "Hormones & Corticosteroids",
    "prednisone": "Hormones & Corticosteroids",
    "pregabalin": "CNS & Psychotropics",
    "primaquine": "Antimalarials",
    "primidone": "CNS & Psychotropics",
    "probenecid": "Musculoskeletal",
    "procainamide": "Cardiovascular",
    "procaine": "Anesthetics",
    "procarbazine": "Oncology",
    "prochlorperazine": "Gastrointestinal",
    "progesterone": "Hormones & Corticosteroids",
    "proguanil": "Antimalarials",
    "promethazine": "Antihistamines & Allergy",
    "propafenone": "Cardiovascular",
    "propofol": "Anesthetics",
    "propranolol": "Cardiovascular",
    "propylthiouracil": "Thyroid Therapy",
    "protamine": "Antidotes & Antivenoms",
    "pseudoephedrine": "Respiratory",
    "psyllium": "Gastrointestinal",
    "pyrantel": "Antihelminthics",
    "pyrazinamide": "Anti-Tuberculosis & Leprosy",
    "pyridostigmine": "Anesthetics",
    "pyridoxine": "Vitamins & Minerals",
    "pyrvinium": "Antihelminthics",
    "quetiapine": "CNS & Psychotropics",
    "quinapril": "Cardiovascular",
    "quinidine": "Cardiovascular",
    "quinine": "Antimalarials",
    "rabeprazole": "Gastrointestinal",
    "raloxifene": "Musculoskeletal",
    "raltegravir": "Antiretrovirals",
    "ramipril": "Cardiovascular",
    "ramelteon": "CNS & Psychotropics",
    "ranibizumab": "Ophthalmologicals",
    "ranitidine": "Gastrointestinal",
    "ranolazine": "Cardiovascular",
    "rasagiline": "CNS & Psychotropics",
    "reboxetine": "CNS & Psychotropics",
    "remifentanil": "Anesthetics",
    "repaglinide": "Antidiabetics",
    "reserpine": "Cardiovascular",
    "retapamulin": "Dermatologicals",
    "reteplase": "Cardiovascular",
    "ribavirin": "Antivirals",
    "riboflavin": "Vitamins & Minerals",
    "rifabutin": "Anti-Tuberculosis & Leprosy",
    "rifampicin": "Anti-Tuberculosis & Leprosy",
    "rifaximin": "Antibiotics",
    "riluzole": "CNS & Psychotropics",
    "rimantadine": "Antivirals",
    "rimexolone": "Ophthalmologicals",
    "risedronate": "Musculoskeletal",
    "risperidone": "CNS & Psychotropics",
    "ritonavir": "Antiretrovirals",
    "rituximab": "Oncology",
    "rivaroxaban": "Anticoagulants",
    "rivastigmine": "CNS & Psychotropics",
    "rizatriptan": "CNS & Psychotropics",
    "roflumilast": "Respiratory",
    "ropinirole": "CNS & Psychotropics",
    "ropivacaine": "Anesthetics",
    "rosiglitazone": "Antidiabetics",
    "rosuvastatin": "Cardiovascular",
    "rotigotine": "CNS & Psychotropics",
    "roxithromycin": "Antibiotics",
    "rufinamide": "CNS & Psychotropics",
    "ruxolitinib": "Oncology",
    "sacubitril": "Cardiovascular",
    "safinamide": "CNS & Psychotropics",
    "salbutamol": "Respiratory",
    "salmeterol": "Respiratory",
    "saquinavir": "Antiretrovirals",
    "sargramostim": "Immunostimulants",
    "scopolamine": "Gastrointestinal",
    "selegiline": "CNS & Psychotropics",
    "senna": "Gastrointestinal",
    "sertaconazole": "Dermatologicals",
    "sertraline": "CNS & Psychotropics",
    "sevelamer": "Renal & Dialysis",
    "sevoflurane": "Anesthetics",
    "sildenafil": "Cardiovascular",
    "silodosin": "Urologicals",
    "silver sulfadiazine": "Dermatologicals",
    "simethicone": "Gastrointestinal",
    "simvastatin": "Cardiovascular",
    "sirolimus": "Immunosuppressants",
    "sitagliptin": "Antidiabetics",
    "sodium bicarbonate": "IV Fluids & Electrolytes",
    "sodium chloride": "IV Fluids & Electrolytes",
    "sodium cromoglicate": "Respiratory",
    "sodium fluoride": "Vitamins & Minerals",
    "sodium phenylbutyrate": "Metabolic Disorders",
    "sodium picosulfate": "Gastrointestinal",
    "sodium thiosulfate": "Antidotes & Antivenoms",
    "sodium valproate": "CNS & Psychotropics",
    "solifenacin": "Urologicals",
    "somatostatin": "Gastrointestinal",
    "somatropin": "Hormones & Corticosteroids",
    "sorafenib": "Oncology",
    "sotalol": "Cardiovascular",
    "spironolactone": "Cardiovascular",
    "stavudine": "Antiretrovirals",
    "streptokinase": "Cardiovascular",
    "streptomycin": "Anti-Tuberculosis & Leprosy",
    "strontium ranelate": "Musculoskeletal",
    "succimer": "Antidotes & Antivenoms",
    "sucralfate": "Gastrointestinal",
    "sufentanil": "Anesthetics",
    "sulbactam": "Antibiotics",
    "sulfacetamide": "Ophthalmologicals",
    "sulfadiazine": "Antibiotics",
    "sulfadoxine": "Antimalarials",
    "sulfamethoxazole": "Antibiotics",
    "sulfasalazine": "Gastrointestinal",
    "sulindac": "Analgesics & Anti-inflammatory",
    "sulpiride": "CNS & Psychotropics",
    "sumatriptan": "CNS & Psychotropics",
    "sunitinib": "Oncology",
    "suplatast": "Respiratory",
    "tacrolimus": "Immunosuppressants",
    "tadalafil": "Urologicals",
    "tafluprost": "Ophthalmologicals",
    "talazoparib": "Oncology",
    "tamsulosin": "Urologicals",
    "tapentadol": "Analgesics & Anti-inflammatory",
    "tasimelteon": "CNS & Psychotropics",
    "tazarotene": "Dermatologicals",
    "tegafur": "Oncology",
    "teicoplanin": "Antibiotics",
    "telmisartan": "Cardiovascular",
    "temazepam": "CNS & Psychotropics",
    "temozolomide": "Oncology",
    "temsirolimus": "Oncology",
    "tenecteplase": "Cardiovascular",
    "tenofovir": "Antiretrovirals",
    "tenoxicam": "Analgesics & Anti-inflammatory",
    "terazosin": "Cardiovascular",
    "terbinafine": "Dermatologicals",
    "terbutaline": "Respiratory",
    "terconazole": "Dermatologicals",
    "teriflunomide": "Immunosuppressants",
    "terlipressin": "Cardiovascular",
    "testosterone": "Hormones & Corticosteroids",
    "tetrabenazine": "CNS & Psychotropics",
    "tetracaine": "Anesthetics",
    "tetracycline": "Antibiotics",
    "tetrahydrozoline": "Ophthalmologicals",
    "thalidomide": "Oncology",
    "theophylline": "Respiratory",
    "thiamine": "Vitamins & Minerals",
    "thiopental": "Anesthetics",
    "thioridazine": "CNS & Psychotropics",
    "thiotepa": "Oncology",
    "thyroxine": "Thyroid Therapy",
    "tiaprofenic": "Analgesics & Anti-inflammatory",
    "ticagrelor": "Cardiovascular",
    "ticlopidine": "Cardiovascular",
    "tigecycline": "Antibiotics",
    "timolol": "Ophthalmologicals",
    "tinidazole": "Antibiotics",
    "tinzaparin": "Anticoagulants",
    "tioconazole": "Dermatologicals",
    "tiotropium": "Respiratory",
    "tipranavir": "Antiretrovirals",
    "tirofiban": "Cardiovascular",
    "tizanidine": "Musculoskeletal",
    "tobramycin": "Antibiotics",
    "tocilizumab": "Immunosuppressants",
    "tolbutamide": "Antidiabetics",
    "tolcapone": "CNS & Psychotropics",
    "tolmetin": "Analgesics & Anti-inflammatory",
    "tolperisone": "Musculoskeletal",
    "tolterodine": "Urologicals",
    "topiramate": "CNS & Psychotropics",
    "topotecan": "Oncology",
    "toremifene": "Oncology",
    "torsemide": "Cardiovascular",
    "tramadol": "Analgesics & Anti-inflammatory",
    "trametinib": "Oncology",
    "trandolapril": "Cardiovascular",
    "tranexamic acid": "Hemostatics",
    "trastuzumab": "Oncology",
    "travoprost": "Ophthalmologicals",
    "trazodone": "CNS & Psychotropics",
    "tretinoin": "Dermatologicals",
    "triamcinolone": "Hormones & Corticosteroids",
    "triamterene": "Cardiovascular",
    "triazolam": "CNS & Psychotropics",
    "trichlormethiazide": "Cardiovascular",
    "triclabendazole": "Antihelminthics",
    "tridihexethyl": "Gastrointestinal",
    "trientine": "Antidotes & Antivenoms",
    "trifluoperazine": "CNS & Psychotropics",
    "trifluridine": "Ophthalmologicals",
    "trihexyphenidyl": "CNS & Psychotropics",
    "trimebutine": "Gastrointestinal",
    "trimethoprim": "Antibiotics",
    "trimetrexate": "Oncology",
    "triptorelin": "Oncology",
    "tromethamine": "IV Fluids & Electrolytes",
    "tropicamide": "Ophthalmologicals",
    "trospium": "Urologicals",
    "ulipristal": "Contraceptives",
    "undecenoic acid": "Dermatologicals",
    "unoprostone": "Ophthalmologicals",
    "urapidil": "Cardiovascular",
    "urea": "Dermatologicals",
    "ursodeoxycholic": "Gastrointestinal",
    "ursodiol": "Gastrointestinal",
    "valacyclovir": "Antivirals",
    "valganciclovir": "Antivirals",
    "valproate": "CNS & Psychotropics",
    "valproic acid": "CNS & Psychotropics",
    "valsartan": "Cardiovascular",
    "vamorolone": "Musculoskeletal",
    "vancomycin": "Antibiotics",
    "vardenafil": "Urologicals",
    "vasopressin": "Cardiovascular",
    "vecuronium": "Anesthetics",
    "velaglucerase": "Metabolic Disorders",
    "venlafaxine": "CNS & Psychotropics",
    "verapamil": "Cardiovascular",
    "verteporfin": "Ophthalmologicals",
    "vigabatrin": "CNS & Psychotropics",
    "vilazodone": "CNS & Psychotropics",
    "vildagliptin": "Antidiabetics",
    "vinblastine": "Oncology",
    "vincristine": "Oncology",
    "vindesine": "Oncology",
    "vinorelbine": "Oncology",
    "vismodegib": "Oncology",
    "vitamin": "Vitamins & Minerals",
    "vitamin a": "Vitamins & Minerals",
    "vitamin b": "Vitamins & Minerals",
    "vitamin c": "Vitamins & Minerals",
    "vitamin d": "Vitamins & Minerals",
    "vitamin e": "Vitamins & Minerals",
    "vitamin k": "Vitamins & Minerals",
    "voriconazole": "Systemic Antifungals",
    "vorinostat": "Oncology",
    "vortioxetine": "CNS & Psychotropics",
    "warfarin": "Anticoagulants",
    "xylometazoline": "ENT Preparations",
    "xipamide": "Cardiovascular",
    "zafirlukast": "Respiratory",
    "zalcitabine": "Antiretrovirals",
    "zaleplon": "CNS & Psychotropics",
    "zanamivir": "Antivirals",
    "zidovudine": "Antiretrovirals",
    "zinc": "Vitamins & Minerals",
    "zoledronic": "Musculoskeletal",
    "zolmitriptan": "CNS & Psychotropics",
    "zolpidem": "CNS & Psychotropics",
    "zonisamide": "CNS & Psychotropics",
    "zopiclone": "CNS & Psychotropics",
    "zuclopenthixol": "CNS & Psychotropics",
    # More specific drug entries
    "acitretin": "Dermatologicals",
    "alcaftadine": "Ophthalmologicals",
    "aloe vera": "Dermatologicals",
    "alpha lipoic acid": "Vitamins & Minerals",
    "alphanate": "Hemostatics",
    "aluminium chloride": "Dermatologicals",
    "antazoline": "Ophthalmologicals",
    "arabinoxylan": "Gastrointestinal",
    "avatrombopag": "Hemostatics",
    "benzydamine": "Analgesics & Anti-inflammatory",
    "bepotastine": "Ophthalmologicals",
    "betaxolol": "Ophthalmologicals",
    "betrixaban": "Anticoagulants",
    "bifidobacterium": "Gastrointestinal",
    "butamirate": "Respiratory",
    "butenafine": "Dermatologicals",
    "carbamide peroxide": "ENT Preparations",
    "carbetocin": "Obstetrics & Gynecology",
    "cenobamate": "CNS & Psychotropics",
    "cefaclor": "Antibiotics",
    "cefazolin": "Antibiotics",
    "cefditoren": "Antibiotics",
    "cetrorelix": "Oncology",
    "chloroxylenol": "Dermatologicals",
    "chlorthalidone": "Cardiovascular",
    "ciprofibrate": "Cardiovascular",
    "clioquinol": "Dermatologicals",
    "clobazam": "CNS & Psychotropics",
    "clobetasol": "Dermatologicals",
    "clobetasone": "Dermatologicals",
    "clomiphene": "Hormones & Corticosteroids",
    "crotamiton": "Dermatologicals",
    "cyclizine": "Gastrointestinal",
    "dapoxetine": "CNS & Psychotropics",
    "daprodustat": "Cardiovascular",
    "desonide": "Dermatologicals",
    "desoximetasone": "Dermatologicals",
    "difluprednate": "Ophthalmologicals",
    "diloxanide": "Antibiotics",
    "dobutamine": "Cardiovascular",
    "docusate": "Gastrointestinal",
    "dopamine": "Cardiovascular",
    "drotaverine": "Gastrointestinal",
    "dydrogesterone": "Hormones & Corticosteroids",
    "eluxadoline": "Gastrointestinal",
    "erdosteine": "Respiratory",
    "ertugliflozin": "Antidiabetics",
    "estradiol": "Hormones & Corticosteroids",
    "estriol": "Hormones & Corticosteroids",
    "faropenem": "Antibiotics",
    "fezolinetant": "CNS & Psychotropics",
    "filgotinib": "Immunosuppressants",
    "filgrastim": "Immunostimulants",
    "fluocinolone": "Dermatologicals",
    "fluorometholone": "Ophthalmologicals",
    "folinic acid": "Antidotes & Antivenoms",
    "follicle stimulating": "Hormones & Corticosteroids",
    "gentian violet": "Dermatologicals",
    "gilteritinib": "Oncology",
    "glipizide": "Antidiabetics",
    "halothane": "Anesthetics",
    "heparin": "Anticoagulants",
    "human chorionic": "Hormones & Corticosteroids",
    "hydrogen peroxide": "Dermatologicals",
    "hydroquinone": "Dermatologicals",
    "icosapent": "Cardiovascular",
    "indapamide": "Cardiovascular",
    "lanthanum carbonate": "Renal & Dialysis",
    "lefamulin": "Antibiotics",
    "lenograstim": "Immunostimulants",
    "lercanidipine": "Cardiovascular",
    "linaclotide": "Gastrointestinal",
    "lurbinectedin": "Oncology",
    "lynestrenol": "Hormones & Corticosteroids",
    "mavacamten": "Cardiovascular",
    "mebhydrolin": "Antihistamines & Allergy",
    "meclizine": "Gastrointestinal",
    "mercaptopurine": "Oncology",
    "mesna": "Antidotes & Antivenoms",
    "methenamine": "Antibiotics",
    "metolazone": "Cardiovascular",
    "mirogabalin": "CNS & Psychotropics",
    "moxonidine": "Cardiovascular",
    "nandrolone": "Hormones & Corticosteroids",
    "natamycin": "Ophthalmologicals",
    "nebulizer": "Medical Devices",
    "nikethamide": "CNS & Psychotropics",
    "obeticholic": "Gastrointestinal",
    "oxymorphone": "Analgesics & Anti-inflammatory",
    "oxyphenonium": "Gastrointestinal",
    "pefloxacin": "Antibiotics",
    "pimecrolimus": "Dermatologicals",
    "pipecuronium": "Anesthetics",
    "piracetam": "CNS & Psychotropics",
    "pirfenidone": "Respiratory",
    "probiotic": "Gastrointestinal",
    "procyclidine": "CNS & Psychotropics",
    "rimegepant": "CNS & Psychotropics",
    "risdiplam": "CNS & Psychotropics",
    "roxadustat": "Cardiovascular",
    "rupatadine": "Antihistamines & Allergy",
    "selinexor": "Oncology",
    "snake venom": "Antidotes & Antivenoms",
    "sodium fusidate": "Antibiotics",
    "sotagliflozin": "Antidiabetics",
    "sotorasib": "Oncology",
    "spectinomycin": "Antibiotics",
    "spinosad": "Dermatologicals",
    "suvorexant": "CNS & Psychotropics",
    "tedizolid": "Antibiotics",
    "tegaserod": "Gastrointestinal",
    "tibolone": "Hormones & Corticosteroids",
    "tolfenamic": "Analgesics & Anti-inflammatory",
    "tolnaftate": "Dermatologicals",
    "tolvaptan": "Cardiovascular",
    "tulobuterol": "Respiratory",
    "varenicline": "CNS & Psychotropics",
    "vinpocetine": "CNS & Psychotropics",
    "zavegepant": "CNS & Psychotropics",
    # More specific entries for remaining unclassified
    "clove oil": "Dermatologicals",
    "coagulation factor": "Hemostatics",
    "condom": "Contraceptives",
    "difamilast": "Dermatologicals",
    "difelikefalin": "Dermatologicals",
    "dotinurad": "Musculoskeletal",
    "ethanol": "Dermatologicals",
    "eucalyptol": "Respiratory",
    "glycine": "IV Fluids & Electrolytes",
    "intravenous fat": "IV Fluids & Electrolytes",
    "irrigating solution": "IV Fluids & Electrolytes",
    "isopropyl alcohol": "Dermatologicals",
    "l-asparaginase": "Oncology",
    "l-ornithine": "Gastrointestinal",
    "oral rehydration": "Gastrointestinal",
    "sodium hyaluronate": "Musculoskeletal",
    "trypan blue": "Diagnostic Contrast Media",
    "water for injection": "IV Fluids & Electrolytes",
    "l-asparaginase": "Oncology",
    "l-lysine": "Vitamins & Minerals",
    "multiple micronutrient": "Vitamins & Minerals",
    "mix of trace": "IV Fluids & Electrolytes",
    "topical agents": "Dermatologicals",
    "white soft paraffin": "Dermatologicals",
    "liquid paraffin": "Dermatologicals",
    "triclosan": "Dermatologicals",
    "sodium alginate": "Gastrointestinal",
    "saccharin": "Gastrointestinal",
    # Combination products / special formulations
    "amino acid": "IV Fluids & Electrolytes",
    "balanced salt solution": "Ophthalmologicals",
    "bp monitoring device": "Medical Devices",
    "contraceptive": "Contraceptives",
    "dialysis": "Renal & Dialysis",
    "electrolytes": "IV Fluids & Electrolytes",
    "hemodialysis": "Renal & Dialysis",
    "parenteral nutrition": "IV Fluids & Electrolytes",
    "peritoneal dialysis": "Renal & Dialysis",
    "ringer": "IV Fluids & Electrolytes",
    "tpn": "IV Fluids & Electrolytes",
    "vaccine": "Vaccines & Immunoglobulins",
}

THERAPEUTIC_CATEGORIES = {
    "antibiotic|amoxicill|ampicillin|cefixime|cefuroxime|ceftriaxone|cefotaxime|cefradine|cephalexin|cephradine|"
    "ciprofloxacin|levofloxacin|ofloxacin|moxifloxacin|norfloxacin|"
    "azithromycin|clarithromycin|erythromycin|spiramycin|"
    "doxycycline|tetracycline|"
    "metronidazole|tinidazole|"
    "cloxacillin|flucloxacillin|penicillin|amoxiclav|co-amoxiclav|"
    "nitrofurantoin|sulfamethoxazole|trimethoprim|cotrimoxazole|"
    "clindamycin|linezolid|vancomycin|gentamicin|neomycin|"
    "polymyxin|bacitracin|mupirocin|fusidic|fucidin|"
    "meropenem|imipenem|ertapenem|doripenem|"
    "cefepime|cefoperazone|cefpodoxime|cefdinir|cefadroxil|cefalexin|"
    "roxithromycin|josamycin|dirithromycin|"
    "rifaximin|teicoplanin|tigecycline|tobramycin|fosfomycin|"
    "sulfadiazine|sulfacetamide|mafenide": "Antibiotics",
    "artemether|lumefantrine|artesunate|amodiaquine|mefloquine|"
    "chloroquine|primaquine|quinine|sulfadoxine|pyrimethamine|"
    "dihydroartemisinin|piperaquine|proguanil|atovaquone": "Antimalarials",
    "paracetamol|acetaminophen|ibuprofen|diclofenac|naproxen|indomethacin|"
    "piroxicam|meloxicam|ketorolac|etoricoxib|celecoxib|"
    "acemetacin|mefenamic|flurbiprofen|ketoprofen|"
    "aspirin|acetylsalicylic|tramadol|morphine|codeine|"
    "pethidine|fentanyl|buprenorphine|nalbuphine|"
    "pregabalin|gabapentin|carbamazepine|"
    "sulindac|tenoxicam|tiaprofenic|tolmetin|nimesulide|"
    "oxycodone|pentazocine|tapentadol|parecoxib": "Analgesics & Anti-inflammatory",
    "amlodipine|nifedipine|felodipine|"
    "atenolol|bisoprolol|propranolol|metoprolol|carvedilol|"
    "enalapril|ramipril|lisinopril|captopril|quinapril|trandolapril|perindopril|"
    "losartan|valsartan|telmisartan|irbesartan|candesartan|olmesartan|"
    "simvastatin|atorvastatin|rosuvastatin|pitavastatin|pravastatin|fluvastatin|"
    "digoxin|amiodarone|verapamil|diltiazem|"
    "furosemide|spironolactone|hydrochlorothiazide|"
    "clopidogrel|ticagrelor|warfarin|rivaroxaban|apixaban|edoxaban|"
    "isosorbide|nitroglycerin|glyceryl trinitrate|"
    "methyldopa|hydralazine|prazosin|doxazosin|terazosin|"
    "nebivolol|acebutolol|esmolol|labetalol|"
    "dipyridamole|ticlopidine|prasugrel|cilostazol|"
    "bumetanide|torasemide|amiloride|triamterene|eplerenone|"
    "eztimibe|fenofibrate|gemfibrozil|niacin|"
    "isosorbide dinitrate|isosorbide mononitrate|"
    "alprostadil|iloprost|bosentan|ambrisentan|riociguat|"
    "nicorandil|ivabradine|ranolazine|"
    "sacubitril|nesiritide|milrinone|levosimendan|"
    "metaraminol|midodrine|phenylephrine|"
    "sotalol|flecainide|propafenone|mexiletine|procainamide|quinidine": "Cardiovascular",
    "salbutamol|albuterol|salmeterol|formoterol|budesonide|fluticasone|"
    "beclomethasone|ipratropium|tiotropium|"
    "montelukast|zafirlukast|theophylline|aminophylline|"
    "acetylcysteine|carbocisteine|bromhexine|ambroxol|dextromethorphan|"
    "indacaterol|olodaterol|umeclidinium|glycopyrronium|"
    "beclomethasone|flunisolide|mometasone|ciclesonide|"
    "benzonatate|bambuterol|orciprenaline|pirbuterol|terbutaline|"
    "roflumilast|pranlukast|suplatast|"
    "dornase alfa|ephedrine|pseudoephedrine": "Respiratory",
    "omeprazole|pantoprazole|esomeprazole|lansoprazole|rabeprazole|"
    "ranitidine|famotidine|cimetidine|nizatidine|"
    "domperidone|metoclopramide|ondansetron|granisetron|palonosetron|"
    "loperamide|bismuth|"
    "lactulose|bisacodyl|senna|psyllium|"
    "mesalazine|mesalamine|sulfasalazine|"
    "ursodeoxycholic|pancreatin|creon|"
    "mebeverine|trimebutine|hyoscyamine|scopolamine|"
    "sucralfate|simethicone|carglumic acid|"
    "octreotide|somatostatin|"
    "glycopyrrolate|prochlorperazine|"
    "lubiprostone|sodium picosulfate": "Gastrointestinal",
    "metformin|glibenclamide|glyburide|gliclazide|glimepiride|"
    "insulin|pioglitazone|rosiglitazone|"
    "sitagliptin|vildagliptin|saxagliptin|linagliptin|alogliptin|"
    "empagliflozin|dapagliflozin|canagliflozin|"
    "acarbose|repaglinide|nateglinide|tolbutamide|"
    "dulaglutide|exenatide|liraglutide|semaglutide|lixisenatide": "Antidiabetics",
    "diazepam|lorazepam|alprazolam|clonazepam|nitrazepam|flurazepam|"
    "fluoxetine|sertraline|citalopram|escitalopram|paroxetine|fluvoxamine|"
    "amitriptyline|nortriptyline|clomipramine|imipramine|doxepin|"
    "haloperidol|risperidone|olanzapine|quetiapine|clozapine|"
    "phenytoin|valproate|valproic|levetiracetam|topiramate|lamotrigine|"
    "carbidopa|levodopa|ropinirole|"
    "donepezil|rivastigmine|memantine|"
    "zolpidem|zopiclone|zaleplon|eszopiclone|"
    "aripiprazole|asenapine|paliperidone|lurasidone|ziprasidone|"
    "venlafaxine|duloxetine|desvenlafaxine|milnacipran|levomilnacipran|"
    "trazodone|nefazodone|vilazodone|vortioxetine|"
    "bupropion|mirtazapine|maprotiline|"
    "lacosamide|oxcarbazepine|rufinamide|eslicarbazepine|"
    "amantadine|rasagiline|selegiline|entacapone|tolcapone|"
    "methylphenidate|dexmethylphenidate|atomoxetine|"
    "lithium|carbamazepine|gabapentin|pregabalin|"
    "buspirone|hydroxyzine|"
    "ergotamine|rizatriptan|sumatriptan|naratriptan|zolmitriptan|eletriptan|frovatriptan|"
    "chlorpromazine|trifluoperazine|fluphenazine|perphenazine|thioridazine|"
    "sulpiride|amisulpride|pimozide|"
    "ramelteon|tasimelteon|"
    "tetrabenazine|deutetrabenazine|"
    "idebenone|riluzole|"
    "memanitine|pramipexole|rotigotine|piribedil|"
    "agomelatine|armodafinil|modafinil|"
    "phenobarbital|primidone|ethosuximide|vigabatrin|tiagabine|": "CNS & Psychotropics",
    "cetirizine|loratadine|desloratadine|fexofenadine|"
    "promethazine|chlorpheniramine|diphenhydramine|"
    "ketotifen|sodium cromoglicate|"
    "levocetirizine|bilastine|ebastine|mizolastine|"
    "pheniramine|hydroxyzine": "Antihistamines & Allergy",
    "levothyroxine|thyroxine|liothyronine|"
    "propylthiouracil|carbimazole|methimazole|"
    "iodine|potassium iodide|lugol": "Thyroid Therapy",
    "prednisolone|prednisone|dexamethasone|hydrocortisone|"
    "betamethasone|triamcinolone|fludrocortisone|"
    "estrogen|progesterone|norethisterone|"
    "testosterone|danazol|clomiphene|"
    "methylprednisolone|desmopressin|vasopressin|"
    "terlipressin|somatropin|octreotide|lanreotide|"
    "clomifene|triptorelin|leuprolide|goserelin|"
    "fludrocortisone|deoxycortone": "Hormones & Corticosteroids",
    "vitamin|thiamine|riboflavin|pyridoxine|cyanocobalamin|"
    "folic acid|ferrous|iron|calcium|magnesium|zinc|"
    "multivitamin|multimineral|"
    "cholecalciferol|ergocalciferol|"
    "ascorbic acid|vit c|retinol|vit a|tocopherol|vit e|"
    "mecobalamin|niacin|biotin|pantothenic|"
    "sodium fluoride|phytonadione|"
    "polysaccharide iron|hydroxocobalamin": "Vitamins & Minerals",
    "vaccine|immunoglobulin|tetanus|diphtheria|hepatitis|"
    "polio|measles|rubella|mumps|bcg|"
    "influenza|pneumococcal|meningococcal|rotavirus|"
    "rabies|yellow fever|typhoid|cholera": "Vaccines & Immunoglobulins",
    "clotrimazole|miconazole|ketoconazole|terbinafine|"
    "hydrocortisone|betamethasone|mometasone|"
    "fusidic|mupirocin|neomycin|bacitracin|"
    "calamine|zinc oxide|coal tar|salicylic acid|"
    "acyclovir|povidone|iodine|chlorhexidine|"
    "permethrin|benzyl benzoate|"
    "tretinoin|adapalene|benzoyl|isotretinoin|"
    "ammonium lactate|urea|podophyllotoxin|"
    "tazarotene|sertaconazole|tioconazole|"
    "silver sulfadiazine|retapamulin|"
    "amorolfine|undecenoic|terconazole": "Dermatologicals",
    "timolol|latanoprost|dorzolamide|brinzolamide|"
    "prednisolone eye|dexamethasone eye|"
    "chloramphenicol eye|gentamicin eye|"
    "ciprofloxacin eye|ofloxacin eye|"
    "ketotifen eye|olopatadine|"
    "lubricant eye|tears|hypromellose|carboxymethylcellulose|"
    "bimatoprost|travoprost|tafluprost|unoprostone|"
    "brimonidine|brinzolamide|"
    "nepafenac|ketorolac eye|"
    "natalizumab|ranibizumab|aflibercept|bevacizumab|verteporfin|"
    "cyclosporine eye|tacrolimus eye|"
    "naphazoline|tetrahydrozoline|"
    "balanced salt solution|"
    "tropicamide|cyclopentolate|atropine eye|"
    "trifluridine|ganciclovir eye|"
    "rimexolone|loteprednol|"
    "emedastine|azelastine|epinastine": "Ophthalmologicals",
    "fluticasone nasal|mometasone nasal|budesonide nasal|"
    "oxymetazoline|xylometazoline|saline nasal|"
    "betahistine|cinnarizine|flunarizine": "ENT Preparations",
    "tamoxifen|letrozole|anastrozole|exemestane|"
    "methotrexate|cyclophosphamide|"
    "capecitabine|fluorouracil|carboplatin|oxaliplatin|cisplatin|"
    "doxorubicin|epirubicin|idarubicin|daunorubicin|mitoxantrone|"
    "paclitaxel|docetaxel|nab-paclitaxel|"
    "vincristine|vinblastine|vindesine|vinorelbine|"
    "etoposide|teniposide|irinotecan|topotecan|"
    "bleomycin|dactinomycin|mitomycin|"
    "imatinib|dasatinib|nilotinib|ponatinib|bosutinib|"
    "sorafenib|sunitinib|pazopanib|axitinib|cabozantinib|"
    "erlotinib|gefitinib|afatinib|osimertinib|dacomitinib|"
    "crizotinib|ceritinib|alectinib|brigatinib|lorlatinib|"
    "venetoclax|idelalisib|ibrutinib|acalabrutinib|"
    "lenalidomide|pomalidomide|thalidomide|"
    "azacitidine|decitabine|"
    "bortezomib|carfilzomib|ixazomib|"
    "fludarabine|cladribine|pentostatin|nelarabine|clofarabine|"
    "carmustine|lomustine|busulfan|melphalan|chlorambucil|thiotepa|"
    "hydroxyurea|hydroxycarbamide|"
    "procarbazine|dacarbazine|temozolomide|"
    "trastuzumab|pertuzumab|cetuximab|panitumumab|rituximab|"
    "bevacizumab|ramucirumab|"
    "nivolumab|pembrolizumab|durvalumab|atezolizumab|avelumab|"
    "cytarabine|gemcitabine|pemetrexed|raltitrexed|"
    "bicalutamide|enzalutamide|flutamide|nilutamide|apalutamide|darolutamide|"
    "anastrozole|letrozole|exemestane|fulvestrant|tamoxifen|toremifene|"
    "leuprolide|goserelin|triptorelin|degarelix|"
    "abiraterone|mitotane|"
    "vismodegib|sonidegib|"
    "regorafenib|vandetanib|lenvatinib|neratinib|lapatinib|"
    "panobinostat|vorinostat|belinostat|romidepsin|"
    "temsirolimus|everolimus|"
    "palbociclib|ribociclib|abemaciclib|"
    "trametinib|cobimetinib|binimetinib|"
    "dabrafenib|encorafenib|vemurafenib|"
    "olaparib|niraparib|rucaparib|talazoparib|"
    "bendamustine|treosulfan|"
    "eribulin|ixabepilone|"
    "anagrelide|arsenic trioxide|"
    "asparaginase|pegaspargase|"
    "alentuzumab|brentuximab|gemtuzumab|inotuzumab|polatuzumab|"
    "belantamab|isatuximab|daratumumab|elotuzumab|"
    "sipuleucel|cemiplimab|"
    "binimetinib|encorafenib|": "Oncology",
    "lidocaine|lignocaine|bupivacaine|ropivacaine|"
    "ketamine|propofol|thiopental|sevoflurane|desflurane|isoflurane|"
    "atropine|neostigmine|suxamethonium|vecuronium|rocuronium|"
    "midazolam|diazepam|lorazepam|"
    "remifentanil|alfentanil|sufentanil|fentanyl|"
    "procaine|tetracaine|articaine|"
    "pancuronium|atracurium|cisatracurium|mivacurium|"
    "epinephrine|adrenaline|norepinephrine|vasopressin|"
    "dexmedetomidine|etomidate|"
    "pyridostigmine|edrophonium|"
    "naloxone|flumazenil|sugammadex|dantrolene|"
    "pralidoxime|glycopyrrolate": "Anesthetics",
    "hemodialysis|peritoneal dialysis|dialysis solution|"
    "sevelamer|calcium acetate|calcitriol|paricalcitol|"
    "erythropoietin|epoetin|darbepoetin|methoxy": "Renal & Dialysis",
    "contraceptive|levonorgestrel|ethinylestradiol|"
    "medroxyprogesterone|etonogestrel|"
    "emergency contraceptive|ulipristal|desogestrel|"
    "norgestrel|norethindrone|drospirenone|"
    "cyproterone|dienogest|chlormadinone": "Contraceptives",
    "sodium chloride|ringer|dextrose|glucose|"
    "mannitol|sodium bicarbonate|hartmann|"
    "parenteral nutrition|tpn|"
    "potassium chloride|potassium phosphate|magnesium sulfate|"
    "tromethamine|amino acid|albumin": "IV Fluids & Electrolytes",
    "tenofovir|emtricitabine|efavirenz|nevirapine|"
    "dolutegravir|raltegravir|"
    "zidovudine|lamivudine|abacavir|"
    "lopinavir|ritonavir|atazanavir|darunavir|"
    "etravirine|rilpivirine|doravirine|"
    "maraviroc|enfuvirtide|"
    "fosamprenavir|saquinavir|tipranavir|"
    "delavirdine|zalcitabine|"
    "stavudine|tenofovir alafenamide|"
    "cabotegravir|bictegravir": "Antiretrovirals",
    "fluconazole|itraconazole|voriconazole|posaconazole|isavuconazole|"
    "amphotericin|terbinafine|griseofulvin|nystatin|"
    "caspofungin|anidulafungin|micafungin": "Systemic Antifungals",
    "rifampicin|isoniazid|pyrazinamide|ethambutol|"
    "streptomycin|bedaquiline|delamanid|pretomanid|"
    "multibacillary|paucibacillary|dapsone|clofazimine|"
    "rifabutin|rifapentine|amikacin|kanamycin|"
    "capreomycin|cycloserine|terizidone|"
    "ethionamide|prothionamide|"
    "para-aminosalicylic|linezolid|": "Anti-Tuberculosis & Leprosy",
    "albendazole|mebendazole|pyrantel|levamisole|"
    "ivermectin|praziquantel|niclosamide|"
    "diethylcarbamazine|triclabendazole|"
    "bithionol|pyrvinium": "Antihelminthics",
    "barium|iopamidol|iohexol|diatrizoate|gadolinium|"
    "ioversol|iodixanol|gadobutrol|gadoteric|gadovist|"
    "barium sulfate|gadoterate|gadodiamide|gadopentetate|": "Diagnostic Contrast Media",
    "naloxone|flumazenil|"
    "deferoxamine|deferasirox|deferiprone|"
    "sodium thiosulfate|methylene blue|"
    "digoxin immune|atropine|pralidoxime|"
    "acetylcysteine|fomepizole|"
    "succimer|trientine|penicillamine|icatibant|"
    "protamine|sugammadex": "Antidotes & Antivenoms",
    "vitamin k|phytomenadione|"
    "protamine|tranexamic acid|"
    "etamsylate|aminocaproic acid|"
    "eltrombopag|romiplostim|"
    "desmopressin|factor viii|factor ix|": "Hemostatics",
    "enoxaparin|dalteparin|tinzaparin|nadroparin|certoparin|"
    "fondaparinux|argatroban|bivalirudin|lepirudin|desirudin|"
    "dabigatran|rivaroxaban|apixaban|edoxaban|": "Anticoagulants",
    "oseltamivir|zanamivir|peramivir|baloxavir|"
    "acyclovir|valacyclovir|famciclovir|penciclovir|"
    "ganciclovir|valganciclovir|"
    "ribavirin|interferon|peginterferon|"
    "remdesivir|favipiravir|"
    "entecavir|tenofovir|lamivudine|adefovir|telbivudine|"
    "boceprevir|telaprevir|sofosbuvir|ledipasvir|velpatasvir|"
    "asunaprevir|daclatasvir|glecaprevir|pibrentasvir|"
    "amantadine|rimantadine|"
    "maraviroc|enfuvirtide|"
    "bictegravir|cabotegravir|": "Antivirals",
    "mycophenolate|azathioprine|cyclosporine|ciclosporin|"
    "tacrolimus|sirolimus|everolimus|"
    "basiliximab|daclizumab|"
    "baricitinib|tofacitinib|upadacitinib|"
    "adalimumab|certolizumab|infliximab|etanercept|"
    "tocilizumab|sarilumab|siltuximab|"
    "belimumab|anifrolumab|"
    "teriflunomide|leflunomide|"
    "fingolimod|dimethyl fumarate|glatiramer|"
    "natalizumab|ocrelizumab|rituximab": "Immunosuppressants",
    "alendronic|alendronate|risedronate|zoledronic|ibandronic|"
    "calcitonin|strontium|"
    "colchicine|baclofen|tolperisone|chlorzoxazone|methocarbamol|"
    "tizanidine|cyclobenzaprine|orphenadrine|carisoprodol|"
    "glucosamine|chondroitin|hyaluronic|"
    "febuxostat|allopurinol|probenecid|"
    "dantrolene|quinine|pamidronate|"
    "abobotulinum|incobotulinum|onabotulinum|": "Musculoskeletal",
    "tamsulosin|alfuzosin|silodosin|dutasteride|finasteride|"
    "solifenacin|tolterodine|oxybutynin|fesoterodine|darifenacin|"
    "mirabegron|trospium|flavoxate|propiverine|"
    "potassium citrate|sodium citrate|"
    "phenazopyridine|bethanechol|": "Urologicals",
    "dinoprostone|misoprostol|oxytocin|ergometrine|"
    "carboprost|sulprostone|gemeprost|"
    "atosiban|ritodrine|fenoterol|": "Obstetrics & Gynecology",
    "gestrinone|danazol|triptorelin|": "Gynecology",
    "idursulfase|velaglucerase|imiglucerase|miglustat|"
    "sodium phenylbutyrate|glycerol phenylbutyrate|"
    "levocarnitine|betaine|carglumic|"
    "nitisinone|sapropterin|": "Metabolic Disorders",
}

# Common drug suffixes that indicate therapeutic class
DRUG_SUFFIX_MAP = [
    (r"(mab|ximab|zumab|umab|monab)$", "Oncology"),
    (r"(nib|tinib)$", "Oncology"),
    (r"(ciclib|parib)$", "Oncology"),
    (r"(lisib)$", "Oncology"),
    (r"vir$", "Antivirals"),
    (r"oxacin$", "Antibiotics"),
    (r"cycline$", "Antibiotics"),
    (r"azole$", "Systemic Antifungals"),
    (r"afil$", "Urologicals"),
    (r"triptan$", "CNS & Psychotropics"),
    (r"prazole$", "Gastrointestinal"),
    (r"vastatin$", "Cardiovascular"),
    (r"sartan$", "Cardiovascular"),
    (r"pril$", "Cardiovascular"),
    (r"olol$", "Cardiovascular"),
    (r"dipine$", "Cardiovascular"),
    (r"zepam$", "CNS & Psychotropics"),
]

DOSAGE_FORM_MAP = {
    "tab": "Tablet", "tablet": "Tablet",
    "cap": "Capsule", "capsule": "Capsule",
    "inj": "Injection", "injection": "Injection", "injectable": "Injection",
    "syr": "Syrup", "syrup": "Syrup",
    "susp": "Suspension", "suspension": "Suspension",
    "cream": "Cream", "ointment": "Ointment", "gel": "Gel",
    "drops": "Drops", "drop": "Drops",
    "eye drop": "Eye Drops", "ear drop": "Ear Drops",
    "nasal spray": "Nasal Spray", "spray": "Spray",
    "inhaler": "Inhaler", "inhalation": "Inhalation",
    "patch": "Patch", "transdermal": "Patch",
    "suppository": "Suppository", "supp": "Suppository",
    "iv infusion": "IV Infusion", "infusion": "IV Infusion",
    "soln": "Solution", "solution": "Solution",
    "powder": "Powder", "powd": "Powder",
    "lotion": "Lotion",
    "shampoo": "Shampoo",
    "paste": "Paste",
    "wafer": "Wafer",
    "dialysis": "Dialysis Solution",
    "granule": "Granules", "granules": "Granules",
    "kit": "Kit",
    "implant": "Implant",
    "pessary": "Pessary",
    "sachet": "Sachet",
    "effervescent": "Effervescent Tablet",
    "dispersible": "Dispersible Tablet",
    "chewable": "Chewable Tablet",
    "film": "Film",
    "lozenge": "Lozenge",
}

# ─── Humanitarian Item Category Classification ────────────────────────────

HUMANITARIAN_CATEGORIES = [
    (r"(blanket|fleece|wool.*blanket)", "Shelter & NFI", "supply"),
    (r"(shade net|plastic sheeting|tarpaulin|tarp)", "Shelter & NFI", "supply"),
    (r"(mosquito net|insecticide.*net|llin|itn)", "Shelter & NFI", "supply"),
    (r"(chlorine|nadcc|water.*treatment|water.*filter|water.*purif)", "WASH", "supply"),
    (r"(dosing pump|submersible pump|water pump|manual pump|syphon)", "WASH", "equipment"),
    (r"(jerrycan|jerry can|water.*container|bladder.*tank)", "WASH", "supply"),
    (r"(soap.*bar|hygiene cap|toilet paper|washing machine)", "WASH", "supply"),
    (r"(bucket|jug|collapsible.*container)", "WASH", "supply"),
    (r"(fire ciment|vermiculite|sandwich panel)", "Shelter & NFI", "supply"),
    (r"(helmet|safety.*jacket|rain.*jacket|dust coat|safety.*overshoe|half.mask|filter p\d)", "Safety & PPE", "supply"),
    (r"(fire blanket|boundary net)", "Safety & PPE", "supply"),
    (r"(shovel|chisel|mallet|hammer|screwdriver|glue gun)", "Tools & Hardware", "equipment"),
    (r"(pallet truck|rack.*heavy|racking)", "Warehouse & Logistics", "equipment"),
    (r"(rope|polypropylene)", "Warehouse & Logistics", "supply"),
    (r"(bag.*food grade|bag.*polypropylene)", "Warehouse & Logistics", "supply"),
    (r"(label.*handling|fragile|this side up|heavy load)", "Warehouse & Logistics", "supply"),
    (r"(hf.*transceiver|hf.*antenna|hf.*tuner|hf.*codan|hf.*icom)", "IT & Communications", "equipment"),
    (r"(vhf.*icom|vhf.*repeater|vhf.*transceiver|kenwood.*tm271)", "IT & Communications", "equipment"),
    (r"(satellite phone|inmarsat|thuraya|isatphone)", "IT & Communications", "equipment"),
    (r"(keyboard.*mouse|wireless.*access point|fortiap)", "IT & Communications", "equipment"),
    (r"(bicycle)", "Transport & Vehicles", "asset"),
    (r"(solar.*panel|solar.*regulator|solar.*light|solar.*system)", "Energy & Power", "equipment"),
    (r"(inverter|charger.inverter|victron|multiplus|quattro|bluesmart|centaur|easy solar)", "Energy & Power", "equipment"),
    (r"(battery.*station|power station|goal zero|bluetti)", "Energy & Power", "equipment"),
    (r"(battery.*optima|battery.*agm|battery.*stationary|battery.*charger)", "Energy & Power", "equipment"),
    (r"(generator.*diesel|fg wilson|wilson.*perkins)", "Energy & Power", "equipment"),
    (r"(circuit breaker|mcb|rccb|main breaker|busbar|distribution box|cable gland|busbar chamber)", "Electrical", "equipment"),
    (r"(cable shoe|fuse|terminal.*battery|bulb.*head|bulb.*fog|bulb.*indicator|spark plug)", "Vehicle Parts", "supply"),
    (r"(lightning.surge|voltage limiter|surge protection)", "Electrical", "equipment"),
    (r"(tl tube|fl tube|tube fixture|light.*fitting)", "Electrical", "equipment"),
    (r"(oscilloscope)", "Instruments & Test Equipment", "equipment"),
    (r"(hfj78|hzj79|hzj7|land cruiser|landcruiser|hilux|kun15|kun25|lan125|lan15|hiace|kdh|lh202|gun125|"
     r"avanza.*f65|massey.*ferguson|yamaha.*ag\d00|grizzly)", "Vehicle Parts", "supply"),
    (r"(brake.*pad|brake.*shoe|brake.*disc|brake.*drum|brake.*cylinder|brake.*hose|brake.*caliper|"
     r"clutch.*disc|clutch.*cover|clutch.*bearing|clutch.*master|clutch.*cylinder)", "Vehicle Parts", "supply"),
    (r"(oil filter|fuel filter|air filter|oil seal|gasket|v-belt|timing belt|fan belt)", "Vehicle Parts", "supply"),
    (r"(shock absorber|stabilizer bar|leading arm|panhard rod|coil spring|leaf spring|bumper.*spring)", "Vehicle Parts", "supply"),
    (r"(tie rod|steering.*knuckle|steering.*damper|steering.*relay|steering.*wheel)", "Vehicle Parts", "supply"),
    (r"(wheel.*bearing|wheel.*hub|wheel.*nut|wheel.*stud|rim|tyre|tire)", "Vehicle Parts", "supply"),
    (r"(headlamp|headlight|combination lamp|side mirror|rear.*view.*mirror|wiper.*blade|wiper.*arm)", "Vehicle Parts", "supply"),
    (r"(starter.*motor|alternator|regulator.*generator|brush.*holder)", "Vehicle Parts", "supply"),
    (r"(injector.*assy|fuel.*pump|fuel.*filter|fuel.*tank|fuel.*level.*sensor)", "Vehicle Parts", "supply"),
    (r"(radiator|water pump|thermostat|fan.*hzj|cool.*air.*duct|grille.*radiator)", "Vehicle Parts", "supply"),
    (r"(trailer|tractor)", "Transport & Vehicles", "asset"),
    (r"(washing machine|refrigerator|cabinet fan)", "Appliances", "equipment"),
    (r"(carton|pallet|stretch hood)", "Warehouse & Logistics", "supply"),
]

# ─── Helper Functions ─────────────────────────────────────────────────────

def classify_drug(generic_name: str) -> tuple[str, str, str]:
    """Returns (therapeutic_category, item_type, is_controlled)"""
    name_lower = generic_name.lower().strip()

    # Extract base drug name by stripping parenthetical qualifiers
    base_name = re.sub(r'\s*\(.*?\)\s*', ' ', name_lower).strip()
    base_name = re.sub(r'\s*\[.*?\]\s*', ' ', base_name).strip()
    base_name = re.sub(r'\s+', ' ', base_name)

    # Step 1: Check explicit mapping first (most specific)
    # Try exact match on full name, then on base name, then partial match
    if name_lower in EXPLICIT_DRUG_MAP:
        category = EXPLICIT_DRUG_MAP[name_lower]
    elif base_name in EXPLICIT_DRUG_MAP:
        category = EXPLICIT_DRUG_MAP[base_name]
    else:
        # Check if any key in the explicit map is a substring of the name
        category = None
        for key, cat in EXPLICIT_DRUG_MAP.items():
            if key in name_lower or key in base_name:
                category = cat
                break

    if category:
        controlled = "TRUE" if any(w in name_lower for w in ["morphine", "fentanyl", "pethidine",
                "buprenorphine", "ketamine", "diazepam", "lorazepam",
                "alprazolam", "clonazepam", "phenobarbital",
                "methylphenidate", "dexamphetamine"]) else "FALSE"
        essential = "TRUE" if any(ed in name_lower for ed in [
                "amoxicill", "paracetamol", "metformin", "salbutamol",
                "omeprazole", "amlodipine", "enalapril", "atenolol",
                "simvastatin", "furosemide", "ceftriaxone", "ciprofloxacin",
                "methotrexate", "artemether", "lumefantrine", "artesunate", "morphine",
                "prednisolone", "diazepam", "insulin", "warfarin",
                "clopidogrel", "ibuprofen", "diclofenac", "ranitidine",
                "metoclopramide", "ondansetron", "fluoxetine", "amitriptyline",
                "haloperidol", "phenytoin", "valproate", "fluconazole",
                "acyclovir", "clotrimazole", "cetirizine", "albendazole",
                "mebendazole", "ivermectin", "gentamicin", "rifampicin", "isoniazid",
                "pyrazinamide", "ethambutol", "doxycycline",
                "azithromycin", "ciprofloxacin", "amoxicillin",
                "hydrochlorothiazide", "losartan", "enalapril",
                "metformin", "insulin", "salbutamol",
        ]) else "FALSE"
        return category, "drug", controlled

    # Step 2: Check suffix-based mapping (for oncology drugs ending in -ib, -mab, etc.)
    for pattern, category in DRUG_SUFFIX_MAP:
        if re.search(pattern, base_name):
            controlled = "FALSE"
            essential = "FALSE"
            return category, "drug", controlled

    # Step 3: Check regex patterns on full name
    for raw_pattern, category in THERAPEUTIC_CATEGORIES.items():
        # Strip trailing | to prevent empty matches
        pattern = raw_pattern.rstrip('|')
        if re.search(pattern, name_lower) or re.search(pattern, base_name):
            # Determine if controlled
            controlled = "FALSE"
            if any(w in name_lower for w in ["morphine", "fentanyl", "pethidine",
                    "buprenorphine", "ketamine", "diazepam", "lorazepam",
                    "alprazolam", "clonazepam", "phenobarbital",
                    "methylphenidate", "dexamphetamine"]):
                controlled = "TRUE"

            # Determine if essential
            essential = "FALSE"
            essential_drugs = [
                "amoxicill", "paracetamol", "metformin", "salbutamol",
                "omeprazole", "amlodipine", "enalapril", "atenolol",
                "simvastatin", "furosemide", "ceftriaxone", "ciprofloxacin",
                "artemether", "lumefantrine", "artesunate", "morphine",
                "prednisolone", "diazepam", "insulin", "warfarin",
                "clopidogrel", "ibuprofen", "diclofenac", "ranitidine",
                "metoclopramide", "ondansetron", "fluoxetine", "amitriptyline",
                "haloperidol", "phenytoin", "valproate", "fluconazole",
                "acyclovir", "clotrimazole", "cetirizine", "albendazole",
                "mebendazole", "ivermectin", "gentamicin",
            ]
            if any(ed in name_lower for ed in essential_drugs):
                essential = "TRUE"

            return category, "drug", controlled

    return "Unclassified Medicines", "drug", "FALSE"


def classify_humanitarian(description: str) -> tuple[str, str, str]:
    """Returns (category, item_type) from description"""
    desc_lower = description.lower()
    for pattern, category, item_type in HUMANITARIAN_CATEGORIES:
        if re.search(pattern, desc_lower):
            return category, item_type
    return "General Supplies", "supply"


def normalize_asset_item(item_name: str) -> tuple[str, str, str]:
    """Normalize asset item name → (clean_name, category, subcategory)"""
    name = item_name.strip()

    # Remove common prefixes
    if name.startswith("Furniture-"):
        clean = name.replace("Furniture-", "").strip()
        subcat = clean.split(",")[0].strip()
        if subcat in ("Generator", "Fans", "Fridge", "Heavy Machinery"):
            if subcat == "Generator":
                return f"Generator {clean}", "Energy & Power", "equipment"
            if subcat == "Fans":
                return f"Fan {clean}", "Appliances", "equipment"
            if subcat == "Fridge":
                return f"Refrigerator {clean}", "Appliances", "equipment"
            if "Heavy" in subcat:
                return f"{clean}", "Equipment & Machinery", "equipment"
        if subcat in ("Advertising", "Branding"):
            return f"{clean}", "Office & Admin", "supply"
        return f"{clean}", "Furniture & Fixtures", "asset"

    if name.startswith("IT-"):
        clean = name.replace("IT-", "").strip()
        categories_map = {
            "Printer": "IT Equipment",
            "Scanner": "IT Equipment",
            "Monitor": "IT Equipment",
            "WiFi": "IT Equipment",
            "Projector": "IT Equipment",
            "Router": "IT Equipment",
            "Laptop": "IT Equipment",
            "Computer": "IT Equipment",
            "Desktop": "IT Equipment",
            "Server": "IT Equipment",
            "Camera": "IT Equipment",
            "Tablet": "IT Equipment",
            "Phone": "Telecommunications",
            "Others": "IT Equipment",
        }
        matched = "IT Equipment"
        for key, cat in categories_map.items():
            if key.lower() in clean.lower():
                matched = cat
                break
        return f"{clean}", matched, "asset"

    # Handle other prefixes
    for prefix in ["Vehicle-", "Machinery-", "Equipment-", "Safety-"]:
        if name.startswith(prefix):
            clean = name.replace(prefix, "").strip()
            cat_map = {
                "Vehicle-": "Transport & Vehicles",
                "Machinery-": "Equipment & Machinery",
                "Equipment-": "Equipment & Machinery",
                "Safety-": "Safety & PPE",
            }
            return f"{clean}", cat_map[prefix], "asset"

    return name, "General Assets", "asset"


def normalize_dosage_form(dosage: str) -> str:
    """Normalize dosage form string"""
    d = dosage.strip().lower()
    if d in DOSAGE_FORM_MAP:
        return DOSAGE_FORM_MAP[d]
    if d and d != "n/a" and d != "na":
        return dosage.strip().title()
    return ""


def make_sku(prefix: str, name: str, idx: int) -> str:
    """Generate SKU from name"""
    safe = re.sub(r'[^a-zA-Z0-9]+', '-', name.strip().lower())
    safe = re.sub(r'-+', '-', safe).strip('-')[:40]
    return f"{prefix}-{safe}-{idx:04d}"


def safe_str(val: str) -> str:
    """Escape and quote CSV value"""
    if not val:
        return ""
    # Only quote if needed
    if ',' in val or '"' in val or '\n' in val:
        return '"' + val.replace('"', '""') + '"'
    return val


def uom_for_item(item_type: str, category: str, name: str) -> str:
    """Guess appropriate UoM"""
    name_lower = name.lower()
    if any(w in name_lower for w in ["tablet", "capsule", "tab", "cap"]):
        return "EA"
    if "sachet" in name_lower or "packet" in name_lower:
        return "PK"
    if any(w in name_lower for w in ["box", "case", "carton"]):
        return "BX"
    if "roll" in name_lower or "tube" in name_lower:
        return "RL"
    if "liter" in name_lower or "litre" in name_lower or "l " in name_lower:
        return "L"
    if any(w in name_lower for w in ["kg", "kilogram", "bag"]):
        return "KG"
    if any(w in name_lower for w in ["metre", "meter", "m "]):  # careful with M
        return "M"
    if item_type in ("asset", "equipment"):
        return "EA"
    if item_type == "drug":
        return "EA"
    return "EA"


# ─── Processing Functions ────────────────────────────────────────────────

def process_drug_files():
    """Process the 3 LATEST_DRUG_LIST files and return classified drug catalogue"""
    drugs_file = os.path.join(DOCS_DIR, "LATEST_DRUG_LIST_BANGLADESH.csv")
    brands_file = os.path.join(DOCS_DIR, "LATEST_DRUG_LIST_BANGLADESH111.csv")

    # Step 1: Read generic drugs
    generics = {}
    with open(drugs_file, "r", encoding="utf-8") as f:
        reader = csv.DictReader(f)
        for row in reader:
            name = row.get("GENERIC_NAME", "").strip()
            url = row.get("GENERIC_URL", "").strip()
            if name:
                generics[name] = {"url": url, "brands": []}

    print(f"  Loaded {len(generics)} generic drug names")

    # Step 2: Read brands and link to generics
    brand_count = 0
    with open(brands_file, "r", encoding="utf-8") as f:
        reader = csv.DictReader(f)
        for row in reader:
            try:
                brand_url = row.get("AVAILABLE_BRAND_URL", "").strip()
                brand_name = row.get("Brand Name", "").strip()
                dosage_form = row.get("Dosage Form", "").strip()
                strength = row.get("Strength", "").strip()
                company = row.get("Company", "").strip()

                # Clean up leading spaces/commas in brand name
                brand_name = brand_name.lstrip(", ").strip()

                if not brand_name:
                    continue

                # Extract generic URL from brand URL
                # URL pattern: https://medex.com.bd/generics/{id}/generic-name/brand-names
                generic_url_match = re.match(r"(https://medex\.com\.bd/generics/\d+/[^/]+)", brand_url)
                if generic_url_match:
                    generic_url = generic_url_match.group(1)
                    # Find matching generic
                    for gname, gdata in generics.items():
                        if gdata["url"] == generic_url:
                            gdata["brands"].append({
                                "name": brand_name,
                                "dosage_form": normalize_dosage_form(dosage_form),
                                "strength": strength if strength.lower() not in ("n/a", "na", "") else "",
                                "company": company,
                            })
                            brand_count += 1
                            break
            except Exception as e:
                print(f"  Warning: Error processing row: {e}", file=sys.stderr)

    print(f"  Linked {brand_count} brand entries to generics")

    # Step 3: Build catalogue entries
    catalogue = []
    idx = 0

    for gname, gdata in sorted(generics.items()):
        if gname == "GENERIC_NAME":
            continue
        if not gname:
            continue

        idx += 1
        category, item_type, controlled = classify_drug(gname)
        is_essential = "TRUE" if any(ed in gname.lower() for ed in [
            "amoxicill", "paracetamol", "metformin", "salbutamol",
            "omeprazole", "amlodipine", "enalapril", "atenolol",
            "simvastatin", "furosemide", "ceftriaxone", "ciprofloxacin",
            "artemether", "lumefantrine", "morphine", "prednisolone",
            "diazepam", "insulin", "ibuprofen", "diclofenac",
            "albendazole", "mebendazole", "ivermectin",
        ]) else "FALSE"

        # Collect unique dosage forms and strengths
        dosage_forms = set()
        strengths = set()
        companies = set()
        for b in gdata["brands"]:
            if b["dosage_form"]:
                dosage_forms.add(b["dosage_form"])
            if b["strength"]:
                strengths.add(b["strength"])
            if b["company"]:
                companies.add(b["company"])

        # Build description
        desc_parts = [gname]
        if strengths:
            desc_parts.append(f"({', '.join(sorted(strengths)[:3])})")
        if dosage_forms:
            desc_parts.append(f"Dosage: {', '.join(sorted(dosage_forms)[:3])}")
        if companies:
            desc_parts.append(f"Mfr: {', '.join(sorted(companies)[:2])}")
        description = " | ".join(desc_parts)

        # Build strengths column
        strength_str = "; ".join(sorted(strengths)) if strengths else ""
        dosage_forms_str = "; ".join(sorted(dosage_forms)) if dosage_forms else ""

        sku = make_sku("DRUG", gname, idx)

        # Determine batch/expiry tracking
        is_batch = "TRUE"
        is_expiry = "TRUE"
        is_cold = "FALSE"
        is_hazard = controlled  # controlled substances are hazardous

        entry = {
            "sku": sku,
            "name": gname,
            "category_name": category,
            "item_type": item_type,
            "description": description,
            "strength": strength_str,
            "dosage_form": dosage_forms_str,
            "uom_abbreviation": "EA",
            "is_batch_tracked": is_batch,
            "is_expiry_tracked": is_expiry,
            "is_cold_chain": is_cold,
            "is_hazardous": is_hazard,
            "is_essential": is_essential,
            "is_controlled": controlled,
            "is_asset": "FALSE",
            "replenishment_type": "min_max",
            "valuation_method": "fifo",
            "manufacturers": "; ".join(sorted(companies)) if companies else "",
            "source": "medex.com.bd",
        }
        catalogue.append(entry)

    return catalogue


def process_asset_file():
    """Process list.csv → asset catalogue"""
    asset_file = os.path.join(DOCS_DIR, "list.csv")
    catalogue = []
    idx = 0
    seen = set()

    with open(asset_file, "r", encoding="utf-8") as f:
        content = f.read()

    lines = content.split("\n")
    header_found = False
    col_names = []

    for line in lines:
        line = line.strip()
        if not line:
            continue

        # Find header row
        if line.startswith("ITEM NAME") and "CATEGORY" in line:
            header_found = True
            parts = line.split(",")
            col_names = [p.strip() for p in parts]
            continue

        if not header_found:
            continue

        parts = line.split(",")
        if len(parts) < 3:
            continue

        item_name = parts[0].strip()
        category_raw = parts[1].strip()
        brand = parts[3].strip() if len(parts) > 3 else ""
        vendor = parts[4].strip() if len(parts) > 4 else ""

        if not item_name or item_name == "ITEM NAME":
            continue

        # Deduplicate
        dedup_key = f"{item_name}|{brand}".lower()
        if dedup_key in seen:
            continue
        seen.add(dedup_key)

        idx += 1
        clean_name, norm_cat, item_type = normalize_asset_item(item_name)

        # Build display name
        display_name = clean_name
        if brand:
            display_name = f"{clean_name} ({brand})"

        sku = make_sku("AST", clean_name, idx)

        description = f"{clean_name}"
        if brand:
            description += f" - Brand: {brand}"
        if vendor:
            description += f" - Vendor: {vendor}"

        entry = {
            "sku": sku,
            "name": display_name,
            "category_name": norm_cat,
            "item_type": item_type,
            "description": description,
            "strength": "",
            "dosage_form": "",
            "uom_abbreviation": "EA",
            "is_batch_tracked": "FALSE",
            "is_expiry_tracked": "FALSE",
            "is_cold_chain": "FALSE",
            "is_hazardous": "FALSE",
            "is_essential": "FALSE",
            "is_controlled": "FALSE",
            "is_asset": "TRUE",
            "replenishment_type": "one_time",
            "valuation_method": "fifo",
            "manufacturers": brand,
            "source": "asset_inventory",
        }
        catalogue.append(entry)

    return catalogue


def process_humanitarian_file():
    """Process medicineanditems.csv → humanitarian/technical catalogue"""
    items_file = os.path.join(DOCS_DIR, "medicineanditems.csv")
    catalogue = []
    idx = 0
    seen = set()

    with open(items_file, "r", encoding="utf-8") as f:
        reader = csv.reader(f)
        for row in reader:
            if not row or not row[0]:
                continue
            desc = row[0].strip()
            if not desc or desc == "Description":
                continue

            # Clean up description
            desc_clean = desc.strip('"').strip()
            if not desc_clean:
                continue

            # Deduplicate
            dedup_key = desc_clean.lower()[:100]
            if dedup_key in seen:
                continue
            seen.add(dedup_key)

            idx += 1
            category, item_type = classify_humanitarian(desc_clean)

            # Extract base name (remove specifics in parentheses for grouping)
            base_name = re.sub(r'\s*\(.*?\)\s*', ' ', desc_clean).strip()
            base_name = re.sub(r'\s+', ' ', base_name)

            # Determine UoM
            uom = uom_for_item(item_type, category, desc_clean)

            # Generate category code prefix
            cat_map = {
                "WASH": "WSH", "Shelter & NFI": "SHL", "Safety & PPE": "SFT",
                "Tools & Hardware": "TLH", "Warehouse & Logistics": "WHL",
                "IT & Communications": "ITC", "Transport & Vehicles": "TRN",
                "Energy & Power": "ENG", "Electrical": "ELC",
                "Instruments & Test Equipment": "INS", "Vehicle Parts": "VPR",
                "Appliances": "APL", "General Supplies": "GEN",
            }

            cat_prefix = "GEN"
            for key, prefix in cat_map.items():
                if key in category:
                    cat_prefix = prefix
                    break

            sku = make_sku(cat_prefix, base_name, idx)

            # Determine properties
            is_cold = "TRUE" if "refrigerat" in desc_clean.lower() or "cold" in desc_clean.lower() else "FALSE"
            is_hazard = "TRUE" if any(w in desc_clean.lower() for w in ["chlorine", "insecticide",
                        "pirimiphos", "deltamethrin", "hazard", "dangerous"]) else "FALSE"
            is_batch = "TRUE" if item_type == "supply" else "FALSE"
            is_expiry = "TRUE" if any(w in desc_clean.lower() for w in ["chlorine", "filter",
                        "medicine", "expir"]) else "FALSE"

            entry = {
                "sku": sku,
                "name": desc_clean[:120],
                "category_name": category,
                "item_type": item_type,
                "description": desc_clean,
                "strength": "",
                "dosage_form": "",
                "uom_abbreviation": uom,
                "is_batch_tracked": is_batch,
                "is_expiry_tracked": is_expiry,
                "is_cold_chain": is_cold,
                "is_hazardous": is_hazard,
                "is_essential": "FALSE",
                "is_controlled": "FALSE",
                "is_asset": "TRUE" if item_type == "asset" else "FALSE",
                "replenishment_type": "min_max",
                "valuation_method": "fifo",
                "manufacturers": "",
                "source": "logistics_catalogue",
            }
            catalogue.append(entry)

    return catalogue


# ─── Main ─────────────────────────────────────────────────────────────────

def main():
    print("=" * 60)
    print("ZarishLog — Master Product Catalogue Generator")
    print("=" * 60)
    os.makedirs(CONFIG_DIR, exist_ok=True)

    # ── Process Drugs ──
    print("\n[1/3] Processing drug catalogues...")
    drugs = process_drug_files()
    print(f"  → {len(drugs)} drug entries classified")

    # ── Process Assets ──
    print("\n[2/3] Processing asset inventory...")
    assets = process_asset_file()
    print(f"  → {len(assets)} asset entries classified")

    # ── Process Humanitarian Items ──
    print("\n[3/3] Processing humanitarian items...")
    humanitarian = process_humanitarian_file()
    print(f"  → {len(humanitarian)} humanitarian entries classified")

    # ── Merge ──
    all_items = drugs + assets + humanitarian

    # ── Category Summary ──
    print("\n" + "=" * 60)
    print("Category Summary")
    print("=" * 60)
    category_counts = defaultdict(int)
    type_counts = defaultdict(int)
    for item in all_items:
        category_counts[item["category_name"]] += 1
        type_counts[item["item_type"]] += 1

    print("\nBy Item Type:")
    for t, c in sorted(type_counts.items(), key=lambda x: -x[1]):
        print(f"  {t}: {c}")

    print("\nBy Category (top 30):")
    for cat, count in sorted(category_counts.items(), key=lambda x: -x[1])[:30]:
        print(f"  {cat}: {count}")

    # ── Write CSV ──
    csv_path = os.path.join(CONFIG_DIR, "master_product_catalogue.csv")
    print(f"\nWriting consolidated catalogue to {csv_path}...")

    csv_fields = [
        "sku", "name", "category_name", "item_type", "description",
        "strength", "dosage_form", "uom_abbreviation",
        "is_batch_tracked", "is_expiry_tracked", "is_cold_chain", "is_hazardous",
        "is_essential", "is_controlled", "is_asset",
        "replenishment_type", "valuation_method", "manufacturers", "source",
    ]

    with open(csv_path, "w", newline="", encoding="utf-8") as f:
        writer = csv.DictWriter(f, fieldnames=csv_fields, quoting=csv.QUOTE_ALL)
        writer.writeheader()
        for item in all_items:
            writer.writerow(item)

    print(f"  → {len(all_items)} total entries written")

    # ── Write Seed SQL ──
    # We need the organizations and categories to reference
    # For now, generate a seed SQL that can be reviewed and adapted

    sql_path = os.path.realpath(os.path.join(
        os.path.dirname(CONFIG_DIR), "..",
        "packages/data-models/seed/data/006_master_product_seed.sql"))
    os.makedirs(os.path.dirname(sql_path), exist_ok=True)

    print(f"Writing seed SQL to {sql_path}...")
    with open(sql_path, "w", encoding="utf-8") as f:
        f.write("""-- ZarishLog — Master Product Seed Data (Phase 6)
-- Generated from medex.com.bd drug registry, asset inventory, and logistics catalogue
-- Idempotent: uses INSERT ... ON CONFLICT DO NOTHING

BEGIN;

-- Reference: get the default org (e.g. CPI Bangladesh)
DO $$
DECLARE
    v_org_id uuid;
    v_cat_id uuid;
    v_uom_id uuid;
    v_user text := 'system';
BEGIN
    SELECT id INTO v_org_id FROM organizations LIMIT 1;

""")

        # Write categories first
        written_cats = set()
        for item in all_items:
            cat = item["category_name"]
            if cat and cat not in written_cats:
                written_cats.add(cat)

        # We'll create a simpler approach: write just the INSERT statements
        # that the DBA can adapt for actual org/uom references

        f.write("""    -- Insert product catalogue entries
    -- NOTE: Update v_org_id, category UUIDs, and UoM UUIDs as needed
    -- These are template statements for the DBA to finalize

""")

        # Write a batch insert for products
        f.write("    -- Products (sample — full list in master_product_catalogue.csv)\n")
        batch_size = 100
        for i in range(0, len(all_items), batch_size):
            batch = all_items[i:i + batch_size]
            f.write("    INSERT INTO products (\n")
            f.write("        org_id, name, sku, category_id, uom_id, item_type,\n")
            f.write("        description, strength, dosage_form_code,\n")
            f.write("        is_batch_tracked, is_expiry_tracked, is_cold_chain, is_hazardous,\n")
            f.write("        is_essential, is_controlled, is_asset,\n")
            f.write("        replenishment_type, valuation_method,\n")
            f.write("        created_by, updated_by\n")
            f.write("    ) VALUES\n")

            values = []
            for item in batch:
                desc_safe = item["description"].replace("'", "''")[:200]
                name_safe = item["name"].replace("'", "''")[:200]
                values.append(
                    f"        (v_org_id, '{name_safe}', '{item['sku']}', v_cat_id, v_uom_id, "
                    f"'{item['item_type']}', "
                    f"'{desc_safe}', "
                    f"'{item['strength'].replace(chr(39), chr(39)+chr(39))}', "
                    f"'{item['dosage_form'].replace(chr(39), chr(39)+chr(39))}', "
                    f"{item['is_batch_tracked']}, {item['is_expiry_tracked']}, "
                    f"{item['is_cold_chain']}, {item['is_hazardous']}, "
                    f"{item['is_essential']}, {item['is_controlled']}, {item['is_asset']}, "
                    f"'{item['replenishment_type']}', '{item['valuation_method']}', "
                    f"v_user, v_user)"
                )

            f.write(",\n".join(values))
            if i + batch_size < len(all_items):
                f.write(";\n\n")
                f.write("    -- ... continuing ...\n")
            else:
                f.write(";\n")

        f.write("""
END;
$$;

COMMIT;
""")

    print(f"  → Seed SQL written to {sql_path}")
    print(f"  → Note: Seed SQL is a template — DBA must resolve org/category/UoM UUIDs")

    # ── Write Category Summary Report ──
    report_path = os.path.join(CONFIG_DIR, "catalogue_summary.txt")
    with open(report_path, "w", encoding="utf-8") as f:
        f.write("ZarishLog — Master Product Catalogue Summary\n")
        f.write("=" * 50 + "\n\n")
        f.write(f"Total entries: {len(all_items)}\n")
        f.write(f"  Drugs: {len(drugs)}\n")
        f.write(f"  Assets: {len(assets)}\n")
        f.write(f"  Humanitarian/Technical: {len(humanitarian)}\n\n")

        f.write("By Item Type:\n")
        for t, c in sorted(type_counts.items(), key=lambda x: -x[1]):
            f.write(f"  {t}: {c}\n")

        f.write("\nBy Category:\n")
        for cat, count in sorted(category_counts.items(), key=lambda x: -x[1]):
            f.write(f"  {cat}: {count}\n")

    print(f"\n  → Summary written to {report_path}")
    print("\nDone!")


if __name__ == "__main__":
    main()
