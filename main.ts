import {confirm, input, select} from "@inquirer/prompts";
import {parseOrder, formatOrder} from "./text-parse.ts";
import * as zipcode from "zipcodes";

import {fetchListOfStores, fetchStoreMenu} from "./fetch-utils.ts";

const askToRun = async () => {
    return confirm({message: `Again?`});
}

const main = async () => {
    let zip = await input({message: "Enter a Zip Code"});
    zip = zipcode.lookup(zip);

    const storeList = await fetchListOfStores(zip.latitude, zip.longitude);

    if(storeList.length > 0) {
        const selectedStore = await select({
            pageSize: 5,
            message: `Select a store`,
            choices: storeList
        });

        const order = await input({message: `Welcome to Taco Bell, can I take your order?`});
        const data = await fetchStoreMenu(selectedStore); /// {name: string, price: number}

        const orderSummary = parseOrder(order, data);
        console.log(formatOrder(orderSummary));
    }
}

(async () => {
    let runAgain = true;
    while(runAgain) {
        await main();
        runAgain = await askToRun();
    }
})();