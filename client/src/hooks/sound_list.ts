import { useCallback, useEffect, useState } from 'react';
import { Category, Sound, api } from '../api';
import useItemList, { ItemList } from './item_list';

export type SoundList = {
    isLoading: boolean;
    categories: ItemList<Category>;
    sounds: ItemList<Sound>;
    refresh: () => Promise<void>;
};

function useSoundList(guildId?: string) {
    const [isLoading, setIsLoading] = useState(false);
    const categories = useItemList<Category>((a, b) => a.id === b.id);

    const onCompare = (a: Sound, b: Sound) => a.id === b.id
    const onFilter = (filter: string, sound: Sound) => 
        !filter || sound.name.toLowerCase().includes(filter.toLowerCase()) || 
        categories.get().find((category) => category.id === sound.categoryId)?.name.toLowerCase().includes(filter.toLowerCase())
    const sounds = useItemList<Sound>(onCompare, onFilter);

 
    const refresh = useCallback(async () => {
        if (!guildId) return;
        setIsLoading(true);
        const [fetchedCategories, fetchedSounds] =
            await api.listCategoriesAndSounds(guildId);
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
