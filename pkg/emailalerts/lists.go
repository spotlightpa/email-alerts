package emailalerts

var fipsToList = map[string]struct{ Name, ID string }{
	"42001": {
		ID:   "7f4f10c4-da18-44d1-981f-03d7014947fb",
		Name: "Adams County",
	},
	"42003": {
		ID:   "69bc4f62-f4be-4429-b923-004ca9a9cfdf",
		Name: "Allegheny County",
	},
	"42005": {
		ID:   "b53aee02-ac91-4e37-8819-9cd37c45e7e8",
		Name: "Armstrong County",
	},
	"42007": {
		ID:   "3bd8378c-3183-445e-bffb-278e4af2c09c",
		Name: "Beaver County",
	},
	"42009": {
		ID:   "5245c06f-b834-4ec8-9acd-aab29a752008",
		Name: "Bedford County",
	},
	"42011": {
		ID:   "1201c0e7-45ac-4238-a783-9945f039c3ca",
		Name: "Berks County",
	},
	"42013": {
		ID:   "cf10fb56-92ab-4b0a-baff-62ffc37b1b64",
		Name: "Blair County",
	},
	"42015": {
		ID:   "af6a14b9-e8c7-4529-ade5-c9755ccf382d",
		Name: "Bradford County",
	},
	"42017": {
		ID:   "8984cce1-64e6-452f-b6bc-823d003ecc3d",
		Name: "Bucks County",
	},
	"42019": {
		ID:   "162e51c6-2da0-4f86-8401-1a92894e6fde",
		Name: "Butler County",
	},
	"42021": {
		ID:   "4798d94e-3311-498d-a9f8-3907717536a9",
		Name: "Cambria County",
	},
	"42023": {
		ID:   "6adc719b-c0df-47d6-8b95-250294ecef4a",
		Name: "Cameron County",
	},
	"42025": {
		ID:   "68bedb98-5832-4e4a-aa5e-27ac9d3cb0e5",
		Name: "Carbon County",
	},
	"42027": {
		ID:   "d3e3e826-e1e7-4533-b671-e9fdeb6b306a",
		Name: "Centre County",
	},
	"42029": {
		ID:   "c058d286-5f31-4ebc-8a68-6e9cf287520f",
		Name: "Chester County",
	},
	"42031": {
		ID:   "7c0d3888-a5c2-4df7-aca3-715104624d86",
		Name: "Clarion County",
	},
	"42033": {
		ID:   "e3601eee-92dd-491f-a507-8ff2e4e6049f",
		Name: "Clearfield County",
	},
	"42035": {
		ID:   "8cceec9c-5dca-408c-9f0a-cd4e92a1c970",
		Name: "Clinton County",
	},
	"42037": {
		ID:   "98ce97e5-7fc3-4c50-9eaf-a31abe8d0185",
		Name: "Columbia County",
	},
	"42039": {
		ID:   "04dd53a4-ba9a-41de-b1b7-e9c2471376fd",
		Name: "Crawford County",
	},
	"42041": {
		ID:   "da190a9a-8cbd-4a8c-896d-97a11e1508de",
		Name: "Cumberland County",
	},
	"42043": {
		ID:   "0724edae-40a6-48e6-8330-cc06b3c67ede",
		Name: "Dauphin County",
	},
	"42045": {
		ID:   "7a76a9de-f271-438f-949b-4452c123b05b",
		Name: "Delaware County",
	},
	"42047": {
		ID:   "2750d1be-9633-4f75-9f96-f5ff8a3f7752",
		Name: "Elk County",
	},
	"42049": {
		ID:   "9ef9cf61-394e-4968-8ebd-18dc270b0e77",
		Name: "Erie County",
	},
	"42051": {
		ID:   "86362cce-3781-4461-833d-7afbc24f1d86",
		Name: "Fayette County",
	},
	"42053": {
		ID:   "84912f53-a7c7-46ed-ba80-10d4a07e9d48",
		Name: "Forest County",
	},
	"42055": {
		ID:   "043de534-deeb-4cdb-acdf-57997603a721",
		Name: "Franklin County",
	},
	"42057": {
		ID:   "77bd41f7-7611-4895-8233-3ad3f500ca41",
		Name: "Fulton County",
	},
	"42059": {
		ID:   "c1dd107d-809d-4aa2-95d9-f47653debaa5",
		Name: "Greene County",
	},
	"42061": {
		ID:   "290994fd-db99-44c5-af79-f840dfd19492",
		Name: "Huntingdon County",
	},
	"42063": {
		ID:   "476a2cf4-b127-474b-a7b8-8bfb50e5b572",
		Name: "Indiana County",
	},
	"42065": {
		ID:   "5b4861cb-534a-4987-a52e-2b680a3fd067",
		Name: "Jefferson County",
	},
	"42067": {
		ID:   "dbcac5ab-cda4-4971-a196-0381e8e97950",
		Name: "Juniata County",
	},
	"42069": {
		ID:   "0334629e-463c-4d89-a931-5bb90c46e34f",
		Name: "Lackawanna County",
	},
	"42071": {
		ID:   "56e027b6-a042-47c8-8149-98586769f406",
		Name: "Lancaster County",
	},
	"42073": {
		ID:   "1cda971d-3cf2-4789-a241-efae3539d76d",
		Name: "Lawrence County",
	},
	"42075": {
		ID:   "1e18ff8c-0130-409a-b6c0-a3e7d2f01156",
		Name: "Lebanon County",
	},
	"42077": {
		ID:   "18fa8f93-1c67-4552-8c2d-34a6586d8386",
		Name: "Lehigh County",
	},
	"42079": {
		ID:   "1fd8301c-a8f4-46c4-84e1-eac856ce3c19",
		Name: "Luzerne County",
	},
	"42081": {
		ID:   "1efef736-0c1e-403b-85bd-cd5578e70b8c",
		Name: "Lycoming County",
	},
	"42083": {
		ID:   "f1de11db-8aa9-4ff1-a6f9-d03ed5595dd1",
		Name: "McKean County",
	},
	"42085": {
		ID:   "31692660-3991-40cd-935e-23b9afdcdeaa",
		Name: "Mercer County",
	},
	"42087": {
		ID:   "572c0f9a-0af3-47b2-bf0e-0a39d5cebdc3",
		Name: "Mifflin County",
	},
	"42089": {
		ID:   "b9799cf2-35e5-401e-a810-e6ffe7ff913b",
		Name: "Monroe County",
	},
	"42091": {
		ID:   "39c3617b-209a-4f12-ac24-b431e77b4135",
		Name: "Montgomery County",
	},
	"42093": {
		ID:   "7c51156b-520a-495e-83db-87c0881b160e",
		Name: "Montour County",
	},
	"42095": {
		ID:   "bf55eff0-8e9c-4444-97db-3eb0ff28ac2a",
		Name: "Northampton County",
	},
	"42097": {
		ID:   "be687656-7946-49ef-8eb0-546fb96a7507",
		Name: "Northumberland County",
	},
	"42099": {
		ID:   "a57f0d1a-eb18-425b-9171-370eb7a7da12",
		Name: "Perry County",
	},
	"42101": {
		ID:   "ff7a408b-cf57-4725-9d12-26751a33ff5f",
		Name: "Philadelphia County",
	},
	"42103": {
		ID:   "2245aae2-23f6-4bd7-b15a-2eb246221aa1",
		Name: "Pike County",
	},
	"42105": {
		ID:   "855fe52a-00a5-458b-ad0e-b20eefcfba82",
		Name: "Potter County",
	},
	"42107": {
		ID:   "ef1babe8-364b-4c94-995e-727f3d972d84",
		Name: "Schuylkill County",
	},
	"42109": {
		ID:   "ac1b3dc2-a184-483a-b54b-bb5c09afdb13",
		Name: "Snyder County",
	},
	"42111": {
		ID:   "a791a53f-c0ff-48d9-b2ea-36c505a923d3",
		Name: "Somerset County",
	},
	"42113": {
		ID:   "896e8b59-b8b4-4c61-9760-17d2d3e75a3b",
		Name: "Sullivan County",
	},
	"42115": {
		ID:   "998f134c-5448-4985-8a79-03bc42f545ef",
		Name: "Susquehanna County",
	},
	"42117": {
		ID:   "6b9c503e-7bdb-4385-86b4-bbb31082fa38",
		Name: "Tioga County",
	},
	"42119": {
		ID:   "d4fe8c5e-2a9f-4107-ac6e-2f20b5d64e19",
		Name: "Union County",
	},
	"42121": {
		ID:   "1c761692-acb8-4727-b2bd-d31c203209fc",
		Name: "Venango County",
	},
	"42123": {
		ID:   "e3b7bb87-6c3c-4004-9c7e-e3057df6ccc9",
		Name: "Warren County",
	},
	"42125": {
		ID:   "3acee895-002c-4fb2-8bab-f0ea7ca19faa",
		Name: "Washington County",
	},
	"42127": {
		ID:   "bfa9b775-4b31-469b-a3b1-a46daec39f5f",
		Name: "Wayne County",
	},
	"42129": {
		ID:   "b6e2ceb1-ac69-4974-a3bc-ec3b0b78026b",
		Name: "Westmoreland County",
	},
	"42131": {
		ID:   "237c0cb4-2d33-4b7d-9dfe-ac1c641db59f",
		Name: "Wyoming County",
	},
	"42133": {
		ID:   "32e0acb1-9580-4622-b144-f37466185a76",
		Name: "York County",
	},
}

var listToFIPS = map[string]struct{ Name, FIPS string }{
	"7f4f10c4-da18-44d1-981f-03d7014947fb": {
		FIPS: "42001",
		Name: "Adams County",
	},
	"69bc4f62-f4be-4429-b923-004ca9a9cfdf": {
		FIPS: "42003",
		Name: "Allegheny County",
	},
	"b53aee02-ac91-4e37-8819-9cd37c45e7e8": {
		FIPS: "42005",
		Name: "Armstrong County",
	},
	"3bd8378c-3183-445e-bffb-278e4af2c09c": {
		FIPS: "42007",
		Name: "Beaver County",
	},
	"5245c06f-b834-4ec8-9acd-aab29a752008": {
		FIPS: "42009",
		Name: "Bedford County",
	},
	"1201c0e7-45ac-4238-a783-9945f039c3ca": {
		FIPS: "42011",
		Name: "Berks County",
	},
	"cf10fb56-92ab-4b0a-baff-62ffc37b1b64": {
		FIPS: "42013",
		Name: "Blair County",
	},
	"af6a14b9-e8c7-4529-ade5-c9755ccf382d": {
		FIPS: "42015",
		Name: "Bradford County",
	},
	"8984cce1-64e6-452f-b6bc-823d003ecc3d": {
		FIPS: "42017",
		Name: "Bucks County",
	},
	"162e51c6-2da0-4f86-8401-1a92894e6fde": {
		FIPS: "42019",
		Name: "Butler County",
	},
	"4798d94e-3311-498d-a9f8-3907717536a9": {
		FIPS: "42021",
		Name: "Cambria County",
	},
	"6adc719b-c0df-47d6-8b95-250294ecef4a": {
		FIPS: "42023",
		Name: "Cameron County",
	},
	"68bedb98-5832-4e4a-aa5e-27ac9d3cb0e5": {
		FIPS: "42025",
		Name: "Carbon County",
	},
	"d3e3e826-e1e7-4533-b671-e9fdeb6b306a": {
		FIPS: "42027",
		Name: "Centre County",
	},
	"c058d286-5f31-4ebc-8a68-6e9cf287520f": {
		FIPS: "42029",
		Name: "Chester County",
	},
	"7c0d3888-a5c2-4df7-aca3-715104624d86": {
		FIPS: "42031",
		Name: "Clarion County",
	},
	"e3601eee-92dd-491f-a507-8ff2e4e6049f": {
		FIPS: "42033",
		Name: "Clearfield County",
	},
	"8cceec9c-5dca-408c-9f0a-cd4e92a1c970": {
		FIPS: "42035",
		Name: "Clinton County",
	},
	"98ce97e5-7fc3-4c50-9eaf-a31abe8d0185": {
		FIPS: "42037",
		Name: "Columbia County",
	},
	"04dd53a4-ba9a-41de-b1b7-e9c2471376fd": {
		FIPS: "42039",
		Name: "Crawford County",
	},
	"da190a9a-8cbd-4a8c-896d-97a11e1508de": {
		FIPS: "42041",
		Name: "Cumberland County",
	},
	"0724edae-40a6-48e6-8330-cc06b3c67ede": {
		FIPS: "42043",
		Name: "Dauphin County",
	},
	"7a76a9de-f271-438f-949b-4452c123b05b": {
		FIPS: "42045",
		Name: "Delaware County",
	},
	"2750d1be-9633-4f75-9f96-f5ff8a3f7752": {
		FIPS: "42047",
		Name: "Elk County",
	},
	"9ef9cf61-394e-4968-8ebd-18dc270b0e77": {
		FIPS: "42049",
		Name: "Erie County",
	},
	"86362cce-3781-4461-833d-7afbc24f1d86": {
		FIPS: "42051",
		Name: "Fayette County",
	},
	"84912f53-a7c7-46ed-ba80-10d4a07e9d48": {
		FIPS: "42053",
		Name: "Forest County",
	},
	"043de534-deeb-4cdb-acdf-57997603a721": {
		FIPS: "42055",
		Name: "Franklin County",
	},
	"77bd41f7-7611-4895-8233-3ad3f500ca41": {
		FIPS: "42057",
		Name: "Fulton County",
	},
	"c1dd107d-809d-4aa2-95d9-f47653debaa5": {
		FIPS: "42059",
		Name: "Greene County",
	},
	"290994fd-db99-44c5-af79-f840dfd19492": {
		FIPS: "42061",
		Name: "Huntingdon County",
	},
	"476a2cf4-b127-474b-a7b8-8bfb50e5b572": {
		FIPS: "42063",
		Name: "Indiana County",
	},
	"5b4861cb-534a-4987-a52e-2b680a3fd067": {
		FIPS: "42065",
		Name: "Jefferson County",
	},
	"dbcac5ab-cda4-4971-a196-0381e8e97950": {
		FIPS: "42067",
		Name: "Juniata County",
	},
	"0334629e-463c-4d89-a931-5bb90c46e34f": {
		FIPS: "42069",
		Name: "Lackawanna County",
	},
	"56e027b6-a042-47c8-8149-98586769f406": {
		FIPS: "42071",
		Name: "Lancaster County",
	},
	"1cda971d-3cf2-4789-a241-efae3539d76d": {
		FIPS: "42073",
		Name: "Lawrence County",
	},
	"1e18ff8c-0130-409a-b6c0-a3e7d2f01156": {
		FIPS: "42075",
		Name: "Lebanon County",
	},
	"18fa8f93-1c67-4552-8c2d-34a6586d8386": {
		FIPS: "42077",
		Name: "Lehigh County",
	},
	"1fd8301c-a8f4-46c4-84e1-eac856ce3c19": {
		FIPS: "42079",
		Name: "Luzerne County",
	},
	"1efef736-0c1e-403b-85bd-cd5578e70b8c": {
		FIPS: "42081",
		Name: "Lycoming County",
	},
	"f1de11db-8aa9-4ff1-a6f9-d03ed5595dd1": {
		FIPS: "42083",
		Name: "McKean County",
	},
	"31692660-3991-40cd-935e-23b9afdcdeaa": {
		FIPS: "42085",
		Name: "Mercer County",
	},
	"572c0f9a-0af3-47b2-bf0e-0a39d5cebdc3": {
		FIPS: "42087",
		Name: "Mifflin County",
	},
	"b9799cf2-35e5-401e-a810-e6ffe7ff913b": {
		FIPS: "42089",
		Name: "Monroe County",
	},
	"39c3617b-209a-4f12-ac24-b431e77b4135": {
		FIPS: "42091",
		Name: "Montgomery County",
	},
	"7c51156b-520a-495e-83db-87c0881b160e": {
		FIPS: "42093",
		Name: "Montour County",
	},
	"bf55eff0-8e9c-4444-97db-3eb0ff28ac2a": {
		FIPS: "42095",
		Name: "Northampton County",
	},
	"be687656-7946-49ef-8eb0-546fb96a7507": {
		FIPS: "42097",
		Name: "Northumberland County",
	},
	"a57f0d1a-eb18-425b-9171-370eb7a7da12": {
		FIPS: "42099",
		Name: "Perry County",
	},
	"ff7a408b-cf57-4725-9d12-26751a33ff5f": {
		FIPS: "42101",
		Name: "Philadelphia County",
	},
	"2245aae2-23f6-4bd7-b15a-2eb246221aa1": {
		FIPS: "42103",
		Name: "Pike County",
	},
	"855fe52a-00a5-458b-ad0e-b20eefcfba82": {
		FIPS: "42105",
		Name: "Potter County",
	},
	"ef1babe8-364b-4c94-995e-727f3d972d84": {
		FIPS: "42107",
		Name: "Schuylkill County",
	},
	"ac1b3dc2-a184-483a-b54b-bb5c09afdb13": {
		FIPS: "42109",
		Name: "Snyder County",
	},
	"a791a53f-c0ff-48d9-b2ea-36c505a923d3": {
		FIPS: "42111",
		Name: "Somerset County",
	},
	"896e8b59-b8b4-4c61-9760-17d2d3e75a3b": {
		FIPS: "42113",
		Name: "Sullivan County",
	},
	"998f134c-5448-4985-8a79-03bc42f545ef": {
		FIPS: "42115",
		Name: "Susquehanna County",
	},
	"6b9c503e-7bdb-4385-86b4-bbb31082fa38": {
		FIPS: "42117",
		Name: "Tioga County",
	},
	"d4fe8c5e-2a9f-4107-ac6e-2f20b5d64e19": {
		FIPS: "42119",
		Name: "Union County",
	},
	"1c761692-acb8-4727-b2bd-d31c203209fc": {
		FIPS: "42121",
		Name: "Venango County",
	},
	"e3b7bb87-6c3c-4004-9c7e-e3057df6ccc9": {
		FIPS: "42123",
		Name: "Warren County",
	},
	"3acee895-002c-4fb2-8bab-f0ea7ca19faa": {
		FIPS: "42125",
		Name: "Washington County",
	},
	"bfa9b775-4b31-469b-a3b1-a46daec39f5f": {
		FIPS: "42127",
		Name: "Wayne County",
	},
	"b6e2ceb1-ac69-4974-a3bc-ec3b0b78026b": {
		FIPS: "42129",
		Name: "Westmoreland County",
	},
	"237c0cb4-2d33-4b7d-9dfe-ac1c641db59f": {
		FIPS: "42131",
		Name: "Wyoming County",
	},
	"32e0acb1-9580-4622-b144-f37466185a76": {
		FIPS: "42133",
		Name: "York County",
	},
	"5a839eb5-d3bc-4f65-9fbe-283c02762a95": {
		FIPS: "test",
		Name: "Test List",
	},
}
