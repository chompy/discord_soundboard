import { useCallback, useEffect, useState } from 'react';
import { api, Category, Sound } from '../api';

export type SoundList = {
    isLoading: boolean;
    categories: Category[];
    sounds: Sound[];
    refresh: () => void;
    localRefresh: () => void;
};

function useSoundList(guildId?: string) {
    const [isLoading, setIsLoading] = useState(false);
    const [categories, setCategories] = useState<Category[]>([]);
    const [sounds, setSounds] = useState<Sound[]>([]);

    const localRefresh = () => {
        setCategories(Array.from(categories));
        setSounds(Array.from(sounds));
    };

    const refresh = useCallback(async () => {
        if (!guildId) return;
        setIsLoading(true);
        const [fetchedCategories, fetchedSounds] = await Promise.all([
            api.listCategories(guildId),
            api.listSounds(guildId),
        ]);
        setCategories(fetchedCategories);
        setSounds(fetchedSounds);
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
        localRefresh,
    };
}

export default useSoundList;
