function extractQuantities(text: string) {
    const numberWords = {
        'one': 1, 'single': 1,
        'two': 2, 'three': 3, 'four': 4, 'five': 5,
        'six': 6, 'seven': 7, 'eight': 8, 'nine': 9, 'ten': 10,
        'eleven': 11, 'twelve': 12, 'dozen': 12
    };

    const matches = [];

    const dozenPattern = /\b(a|an)\s+dozen\s+([a-z\s]+?)(?=\s+and\b|$|[,?!])/gi;
    let match;
    while ((match = dozenPattern.exec(text)) !== null) {
        const item = match[2].trim().toLowerCase();
        matches.push({quantity: 12, item});
    }

    const digitPattern = /\b(\d+)\s+([a-z\s]+?)(?=\s+and\b|$|[,?!])/gi;
    while ((match = digitPattern.exec(text)) !== null) {
        const quantity = parseInt(match[1]);
        const item = match[2].trim().toLowerCase();
        matches.push({quantity, item});
    }

    const wordPattern = /\b(one|two|three|four|five|six|seven|eight|nine|ten|eleven|twelve|dozen|single)\s+([a-z\s]+?)(?=\s+and\b|$|[,?!])/gi;
    while ((match = wordPattern.exec(text)) !== null) {
        const beforeMatch = text.substring(Math.max(0, match.index - 3), match.index);
        if (beforeMatch.match(/\b(a|an)\s*$/i)) {
            continue;
        }

        const quantity = numberWords[match[1].toLowerCase()];
        const item = match[2].trim().toLowerCase();
        matches.push({quantity, item});
    }

    const indefinitePattern = /\b(a|an)\s+(?!dozen\b)([a-z\s]+?)(?=\s+and\b|$|[,?!])/gi;
    while ((match = indefinitePattern.exec(text)) !== null) {
        const item = match[2].trim().toLowerCase();
        matches.push({quantity: 1, item});
    }

    return matches;
}

function levenshteinDistance(a: string, b: string) {
    const matrix = [];

    for (let i = 0; i <= b.length; i++) {
        matrix[i] = [i];
    }
    for (let j = 0; j <= a.length; j++) {
        matrix[0][j] = j;
    }

    for (let i = 1; i <= b.length; i++) {
        for (let j = 1; j <= a.length; j++) {
            if (b.charAt(i - 1) === a.charAt(j - 1)) {
                matrix[i][j] = matrix[i - 1][j - 1];
            } else {
                matrix[i][j] = Math.min(matrix[i - 1][j - 1] + 1, Math.min(matrix[i][j - 1] + 1, matrix[i - 1][j] + 1));
            }
        }
    }
    return matrix[b.length][a.length];
}

function calculateSimilarity(a: string, b: string) {
    const longerWord = a.length > b.length ? a : b;
    const shorterWord = a.length > b.length ? b : a;

    if(longerWord.length === 0) return 1.0;

    const distance = levenshteinDistance(longerWord, shorterWord);

    return (longerWord.length - distance) / longerWord.length;
}

function findBestMatch(userInput, menuItems) {
    const input = userInput.trim().toLowerCase();
    let bestMatch = null;
    let bestScore = 0;

    const normalizeWord = (word: string) => {
        return word.replace(/s$/, '');
    };

    const normalizedInput = normalizeWord(input);

    for(const item of menuItems) {
        const itemName = item.name.toLowerCase();
        const normalizedItemName = normalizeWord(itemName);

        if(itemName.includes(input) || input.includes(itemName) ||
            normalizedItemName.includes(normalizedInput) || normalizedInput.includes(normalizedItemName)) {
            const score = Math.max(input.length / itemName.length, itemName.length / input.length);

            if(score > bestScore) {
                bestMatch = item;
                bestScore = score;
            }
        }

        const similarity = calculateSimilarity(input, itemName);
        const normalizedSimilarity = calculateSimilarity(normalizedInput, normalizedItemName);

        if(similarity > bestScore && similarity > 0.65) {
            bestMatch = item;
            bestScore = similarity;
        }

        if(normalizedSimilarity > bestScore && normalizedSimilarity > 0.65) {
            bestMatch = item;
            bestScore = normalizedSimilarity;
        }

        const inputWords = input.split(/\s+/).map(normalizeWord);
        const itemWords = itemName.split(/\s+/).map(normalizeWord);
        const wordMatches = inputWords.filter(inputWord =>
            itemWords.some(itemWord =>
                itemWord.includes(inputWord) || inputWord.includes(itemWord) ||
                calculateSimilarity(inputWord, itemWord) > 0.8
            )
        );

        const wordScore = wordMatches.length / Math.max(inputWords.length, itemWords.length);
        if (wordScore > bestScore && wordScore > 0.5) {
            bestScore = wordScore;
            bestMatch = item;
        }
    }

    return bestMatch ? {item: bestMatch, confidence: bestScore} : null;
}

export function parseOrder(orderText: string, menuItems) {
    const quantities = extractQuantities(orderText);
    const results = {
        items: [],
        errors: [],
        total: 0
    };

    if(quantities.length === 0) {
        const match = findBestMatch(orderText, menuItems);

        if(match) {
            results.items.push({
                name: match.item.name,
                price: parseFloat(match.item.price),
                quantity: 1,
                subtotal: parseFloat(match.item.price),
                confidence: match.confidence
            });
            results.total = parseFloat(match.item.price);
        } else {
            results.errors.push(`Could not find a match for "${orderText}".`);
        }
    } else {
        quantities.forEach(({quantity, item}) => {
            const match = findBestMatch(item, menuItems);

            if(match) {
                const price = parseFloat(match.item.price);
                const subtotal = price * quantity;

                results.items.push({
                    name: match.item.name,
                    price,
                    quantity: quantity,
                    subtotal,
                    confidence: match.confidence
                });
                results.total += subtotal;
            } else {
                results.errors.push(`Could not find a match for "${item}".`);
            }
        });
    }

    return results;
}

export function formatOrder(results) {
    let output = [];

    if(results.items.length > 0) {
        output.push("=== YOUR ORDER ===");
        results.items.forEach(item => {
            output.push(`${item.quantity}x ${item.name} - $${item.subtotal.toFixed(2)}`);
            if (item.confidence < 0.8) {
                output.push(`(⚠️ Low confidence match - is this correct?)`);
            }
        });
        output.push(`---`);
        output.push(`TOTAL: $${results.total.toFixed(2)}`);
    }

    if (results.errors.length > 0) {
        output.push("\n=== ERRORS ===");
        results.errors.forEach(error => output.push(`❌ ${error}`));
    }

    return output.join('\n');
}