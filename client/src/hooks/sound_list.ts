import { useEffect, useState } from 'react';
import { api, Category, Sound } from '../api';

function UseSoundList(guildId: string) {
    const [isLoading, setIsLoading] = useState(true);
    const [categories, setCategories] = useState<Category[]>([]);
    const [sounds, setSounds] = useState<Sound[]>([]);

    const refresh = () => {
        Promise.all([
            api.listCategories(guildId),
            api.listSounds(guildId),
        ]).then(([fetchedCategories, fetchedSounds]) => {
            setCategories(fetchedCategories);
            setSounds(fetchedSounds);
            setIsLoading(false);
        });
    };

    useEffect(refresh, []);

    const soundInCategory = (category: Category) =>
        sounds.filter((sound) => sound.categoryId === category.id);

    const updateCategory = (category: Category) => {
        const index = categories.findIndex(
            (iterCategory) => iterCategory.id === category.id
        );
        if (index >= 0) {
            categories[index] = category;
        } else {
            categories.push(category);
        }
        setCategories(Array.from(categories).sort((a, b) => a.sort - b.sort));
    };

    const removeCategory = (category: Category) => {
        const index = categories.findIndex(
            (iterCategory) => iterCategory.id === category.id
        );
        const updatedCategories = Array.from(categories);
        updatedCategories.splice(index, 1);
        setCategories(updatedCategories);
    };

    const updateSound = (sound: Sound, remove?: boolean) => {
        const index = sounds.findIndex(
            (iterSound) => iterSound.id === sound.id
        );
        if (index >= 0) {
            sounds.splice(index, 1);
        }
        setSounds(
            [...sounds, ...(remove ? [] : [sound])].sort(
                (a, b) => a.sort - b.sort
            )
        );
    };

    const removeSound = (sound: Sound) => {
        const index = sounds.findIndex(
            (iterSound) => iterSound.id === sound.id
        );
        const updatedSounds = Array.from(sounds);
        updatedSounds.splice(index, 1);
        setSounds(updatedSounds);
    };

    return {
        isLoading,
        categories,
        sounds,
        soundInCategory,
        updateCategory,
        updateSound,
        removeCategory,
        removeSound,
        refresh,
    };
}

export default UseSoundList;
