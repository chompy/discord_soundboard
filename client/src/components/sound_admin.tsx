import React, { useState, useEffect } from 'react';

import { api, Category, Sound } from '../api';
import Button from './button';
import SoundAdminOption from './sound_admin_option';
import { convertSound } from '../converter';
import { error } from '../utils';

export type SoundAdminProperties = {
    guildId: string;
};

function SoundAdmin({ guildId }: SoundAdminProperties) {
    const [isLoading, setIsLoading] = useState(true);
    const [categories, setCategories] = useState<Category[]>([]);
    const [sounds, setSounds] = useState<Sound[]>([]);
    const [activeCategory, setActiveCategory] = useState<Category | null>(null);

    useEffect(() => {
        Promise.all([
            api.listCategories(guildId),
            api.listSounds(guildId),
        ]).then(([categories, sounds]) => {
            setCategories(categories);
            setSounds(sounds);
            setActiveCategory(categories[0]);
            setIsLoading(false);
        });
    }, []);

    const updateCategory = (category: Category) => {
        const index = categories.findIndex(
            (iterCategory) => iterCategory.id === category.id
        );
        if (index >= 0) {
            categories[index] = category;
        } else {
            categories.push(category);
        }
        categories.sort((a, b) => a.sort - b.sort);
        setCategories(categories);
        setActiveCategory(category);
    };

    const addCategory = async () => {
        if (isLoading) return;
        const name = prompt('Enter category name:')?.trim();
        if (!name) return;
        setIsLoading(true);
        updateCategory(await api.saveCategory({ name, guildId }));
        setIsLoading(false);
    };

    const deleteCategory = async (category: Category) => {
        if (isLoading) return;
        const catSounds = sounds.filter(
            (sound) => sound.categoryId === category.id
        );

        if (
            catSounds.length > 0 &&
            !confirm(
                `Are you sure you want to delete this category? All ${catSounds.length} sound(s) will be deleted as well.`
            )
        ) {
            return;
        }

        setIsLoading(true);
        await api.deleteCategory(category);
        setCategories(
            categories.filter((iterCategory) => iterCategory.id !== category.id)
        );
        if (activeCategory.id === category.id) {
            setActiveCategory(categories.length > 0 ? categories[0] : null);
        }
        setIsLoading(false);
    };

    const renameCategory = async (category: Category) => {
        if (isLoading) return;
        const name = prompt('Enter category name:', category.name)?.trim();
        if (!name) return;
        setIsLoading(true);
        category.name = name;
        updateCategory(await api.saveCategory(category));
        setIsLoading(false);
    };

    const updateSound = (sound: Sound) => {
        if (isLoading) return;
        const index = sounds.findIndex(
            (iterSound) => iterSound.id === sound.id
        );
        if (index >= 0) {
            sound[index] = sound;
        } else {
            sounds.push(sound);
        }
        sounds.sort((a, b) => a.sort - b.sort);
        setSounds(sounds);
    };

    const addSound = async () => {
        if (isLoading) return;
        document.getElementById('sound-admin-file').click();
    };

    const onFile = async () => {
        if (!activeCategory || isLoading) return;

        setIsLoading(true);
        const fileElement = document.getElementById(
            'sound-admin-file'
        ) as HTMLInputElement;
        if (fileElement.files) {
            // TODO try/catch doesn't catch errors??
            let soundData = null;
            try {
                soundData = await convertSound(
                    await fileElement.files[0].arrayBuffer()
                );
            } catch (e) {
                error(`Unable to decode sound: ${e}`);
                setIsLoading(false);
                return;
            }

            const hash = await api.uploadSound(soundData);
            const sound = await api.saveSound({
                name: '(new sound)',
                hash,
                categoryId: activeCategory.id,
            });
            updateSound(sound);
        }
        setIsLoading(false);
    };

    const renameSound = async (sound: Sound) => {
        if (isLoading) return;
        const name = prompt('Enter sound name:', sound.name)?.trim();
        if (!name) return;
        setIsLoading(true);
        sound.name = name;
        updateSound(await api.saveSound(sound));
        setIsLoading(false);
    };

    const deleteSound = async (sound: Sound) => {
        if (isLoading) return;
        setIsLoading(true);
        await api.deleteSound(sound);
        setSounds(sounds.filter((iterSound) => iterSound.id !== sound.id));
        setIsLoading(false);
    };

    return (
        <div className="sound-admin">
            <h1>Sound Admin</h1>
            <div className="options">
                <Button
                    label="Add Category"
                    disabled={isLoading}
                    onClick={addCategory}
                />
                <Button
                    label="Add Sound"
                    disabled={isLoading}
                    onClick={addSound}
                />
            </div>
            <input id="sound-admin-file" type="file" onChange={onFile} />
            <div className="categories">
                <ul>
                    {categories.map((category) => (
                        <SoundAdminOption
                            key={`category-${category.id}`}
                            active={activeCategory.id === category.id}
                            label={category.name}
                            onClick={() => setActiveCategory(category)}
                            onDelete={() => deleteCategory(category)}
                            onEdit={() => renameCategory(category)}
                        />
                    ))}
                </ul>
            </div>
            <div className="sounds">
                <ul>
                    {sounds
                        .filter(
                            (sound) => sound.categoryId === activeCategory.id
                        )
                        .map((sound) => (
                            <SoundAdminOption
                                key={`sound-${sound.id}`}
                                label={sound.name}
                                onEdit={() => renameSound(sound)}
                                onDelete={() => deleteSound(sound)}
                            />
                        ))}
                </ul>
            </div>
        </div>
    );
}

export default SoundAdmin;
