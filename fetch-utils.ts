export const fetchListOfStores = async (lat: number, long: number) => {
    const storeListUrl = `https://www.tacobell.com/tacobellwebservices/v4/tacobell/stores?latitude=${lat}&longitude=${long}`;

    const response = await fetch(storeListUrl);
    const data = await response.json();

    let formattedList = [];

    if(data.nearByStores.length > 0) {
        const storeData = data.nearByStores.length > 5 ? data.nearByStores.slice(0, 5) : data.nearByStores;

        for (const store of storeData) {
            formattedList.push({
                name: `${store.address.line1} ${store.address.town}, ${store.address.region.isocode.substring(3)}`,
                value: store.storeNumber
            });
        }

        return formattedList;
    } else {
        return formattedList;
    }
}

export const fetchStoreMenu = async (storeNumber: number) => {
    const storeMenuUrl = `https://www.tacobell.com/tacobellwebservices/v4/tacobell/products/menu/${storeNumber}`;

    const response = await fetch(storeMenuUrl);
    const data = await response.json();

    return data.menuProductCategories.flatMap(category =>
        category.products
            .filter(product => product.price.value > 0)
            .map(product => ({
                name: formatProductName(product.name),
                price: product.price.value
            }))
    );
}

const formatProductName = (name: string) => {
    return name.replaceAll(/[®™©℠]/g, '').toLowerCase();
}