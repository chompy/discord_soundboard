import { useCallback, useEffect, useState } from 'react';
import { Category, Sound, UserFavorite, UserKeybind, api } from '../api';
import useItemList from './item_list';


function useSoundList(guildId?: string) {
    const [isLoading, setIsLoading] = useState(false);
    const categories = useItemList<Category>((a, b) => a.id === b.id);

    const onCompare = (a: Sound, b: Sound) => a.id === b.id
    const onFilter = (filter: string, sound: Sound) => 
        !filter || sound.name.toLowerCase().includes(filter.toLowerCase()) || 
        categories.get().find((category) => category.id === sound.categoryId)?.name.toLowerCase().includes(filter.toLowerCase())
    const sounds = useItemList<Sound>(onCompare, onFilter);

    const favorites = useItemList<UserFavorite>((a, b) => a.soundId === b.soundId);
    const keybinds = useItemList<UserKeybind>((a, b) => a.soundId === b.soundId);
 
    const refresh = useCallback(async () => {
        if (!guildId) return;
        setIsLoading(true);
        const [fetchedCategories, fetchedSounds] =
            await api.listCategoriesAndSounds(guildId);
        const fetchedUserFavorites = await api.listUserFavorites();
        const fetchedUserKeybinds = await api.listUserKeybinds();
        categories.update(...fetchedCategories);
        sounds.update(...fetchedSounds);
        favorites.update(...fetchedUserFavorites);
        keybinds.update(...fetchedUserKeybinds);
        setIsLoading(false);
    }, [guildId]);

    const setFavorite = useCallback(async (sound: Sound, favorite: boolean) => {;
        if (!favorite) {
            await api.deleteUserFavorite(sound);
            favorites.delete({userId: 0, soundId: sound.id, created: new Date()});
            return;
        }
        await api.addUserFavorite(sound);
        favorites.update({userId: 0, soundId: sound.id, created: new Date()});
    }, [guildId, favorites]);

    const setKeybind = useCallback(async (sound: Sound, key: string | null) => {;
        if (key === null || key === '') {
            await api.deleteUserKeybindForSound(sound);
            keybinds.delete({userId: 0, soundId: sound.id, key: '', created: new Date()});
            return;
        }
        await api.deleteUserKeybindForKey(key)
        await api.addUserKeybind(sound, key);
        keybinds.update({userId: 0, soundId: sound.id, key, created: new Date()}, ...keybinds.get().flatMap((keybind) => keybind.key === key ? [{...keybind, key: null}] : []));
    }, [guildId, keybinds]);

    useEffect(() => {
        refresh();
    }, [guildId]);

    return {
        isLoading,
        categories,
        sounds,
        favorites,
        keybinds,
        refresh,
        setFavorite,
        setKeybind
    };
}

export default useSoundList;

export type SoundList = ReturnType<typeof useSoundList>;