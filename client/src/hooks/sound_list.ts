import { useCallback, useEffect, useState } from 'react';
import { api, Category, Sound } from '../api';
import useItemList, { ItemList } from './item_list';

export type SoundList = {
    isLoading: boolean;
    categories: ItemList<Category>;
    sounds: ItemList<Sound>;
    refresh: () => Promise<void>;
};

function useSoundList(guildId?: string) {
    const [isLoading, setIsLoading] = useState(false);
    const sounds = useItemList<Sound>((a, b) => a.id === b.id);
    const categories = useItemList<Category>((a, b) => a.id === b.id);

    const refresh = useCallback(async () => {
        if (!guildId) return;
        setIsLoading(true);
        const [fetchedCategories, fetchedSounds] = await Promise.all([
            api.listCategories(guildId),
            api.listSounds(guildId),
        ]);
        categories.update(...fetchedCategories);
        sounds.update(...fetchedSounds);
        setIsLoading(false);
    }, [guildId]);

    useEffect(() => {
        refresh();
    }, [guildId]);

    return {
        isLoading,
        categories,
        sounds,
        refresh,
    };
}

export default useSoundList;
