const { Parser } = require("json2csv");
const fs = require("fs");

const data = JSON.parse(fs.readFileSync("clusterInfo.json"));

function flattenSubDocs(json) {
  for (let key in json) {
    if (Array.isArray(json[key])) {
      json[key].forEach((subDoc, index) => {
        for (let subKey in subDoc) {
          json[`${key}_${index}_${subKey}`] = subDoc[subKey];
        }
      });
      delete json[key];
    } else if (typeof json[key] === "object") {
      flattenSubDocs(json[key]);
    }
  }
  return json;
}

const flattenedData = flattenSubDocs(data);

try {
  const opts = {};
  const parser = new Parser(opts);
  const csv = parser.parse(flattenedData);
  fs.writeFileSync("clusterInfo.csv", csv);
} catch (err) {
  console.error(err);
}
