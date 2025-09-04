import { useCallback, useState } from 'react';

export type ItemList<T> = {
    get: () => T[];
    update: (...items: T[]) => T[];
    delete: (...items: T[]) => T[];
    length: () => number;
    moveIndex: (from: number, to: number) => T[];
    move: (from: T, to: T) => T[];
    clear: () => void;
};

function useItemList<T>(onCompare: (a: T, b: T) => boolean) {
    const [items, setItems] = useState<T[]>([]);

    const get = useCallback(() => Array.from(items), [items]);
    const update = useCallback(
        (...items: T[]) => {
            const newItems = get();
            items.forEach((item) => {
                const index = newItems.findIndex((iterItem) =>
                    onCompare(iterItem, item)
                );
                if (index >= 0) newItems[index] = item;
                else newItems.push(item);
            });
            setItems(newItems);
            return newItems;
        },
        [items]
    );
    const deleteItem = useCallback(
        (...items: T[]) => {
            const newItems = get();
            items.forEach((item) => {
                const index = newItems.findIndex((iterItem) =>
                    onCompare(iterItem, item)
                );
                if (index >= 0) newItems.splice(index, 1);
            });
            setItems(newItems);
            return newItems;
        },
        [items]
    );
    const length = useCallback(() => items.length, [items]);

    const moveIndex = useCallback(
        (from: number, to: number) => {
            const newItems = get();
            const item = newItems[from];
            newItems.splice(from, 1);
            newItems.splice(to, 0, item);
            setItems(newItems);
            return newItems;
        },
        [items]
    );

    const move = useCallback(
        (from: T, to: T) => {
            const fromIndex = items.findIndex((item) => onCompare(item, from));
            const toIndex = items.findIndex((item) => onCompare(item, to));
            if (fromIndex >= 0 && toIndex >= 0)
                return moveIndex(fromIndex, toIndex);
            return get();
        },
        [items]
    );

    const clear = () => setItems([]);

    return {
        get,
        update,
        delete: deleteItem,
        length,
        moveIndex,
        move,
        clear,
    };
}

export default useItemList;
