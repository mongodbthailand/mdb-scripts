const fs = require("fs");
async function main() {
  console.log("STARTING....");
  let data = [];
  const dbs = db
    .getMongo()
    .getDBNames()
    .filter((db) => db !== "admin" && db !== "local" && db !== "config");
  for (adb of dbs) {
    const MB = 1024 * 1024;
    const colls = db.getSiblingDB(adb).getCollectionNames();
    for (coll of colls) {
      const info = db.getSiblingDB(adb).getCollection(coll).stats(); //Bytes
      if (info.ok) {
        data.push({
          db: adb,
          ns: info.ns,
          count: info.count,
          sizeMB: parseFloat(info.size / MB),
          storageSizeMB: parseFloat(info.storageSize / MB),
          nindexes: info.nindexes,
          totalIndexSizeMB: parseFloat(info.totalIndexSize / MB),
          indexSizes: await getIndexStats(
            db.getSiblingDB(adb).getCollection(coll)
          ),
        });
      }
    }
  }
  const content = JSON.stringify(data);
  // printjson(content);
  fs.writeFile("clusterInfo.json", content, (err) => {
    if (err) {
      console.error(err);
    }
    console.log("DONE!");
  });
}
//
async function getIndexStats(coll) {
  let arrayOfIndexStats = [];
  const indexStatsResults = await coll
    .aggregate([{ $indexStats: {} }])
    .toArray();
  for (let indexStatsResult of indexStatsResults) {
    arrayOfIndexStats.push({
      key: indexStatsResult.key,
      name: indexStatsResult.name,
      accesses: parseFloat(indexStatsResult.accesses.ops),
      since: indexStatsResult.accesses.since,
    });
  }
  return arrayOfIndexStats;
}
main();
// mongosh [connection string] --quiet mdbstats.js
