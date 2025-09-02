import React, { useState, useEffect, useRef, act } from 'react';

import { api, Category, Sound } from '../api';
import Button from './button';
import SoundAdminOption from './sound_admin_option';
import { sound as soundUtils } from '../sound';
import { error, log } from '../utils';
import SortableList, { SortableItem } from 'react-easy-sort';
import UseSoundList from '../hooks/sound_list';

export type SoundAdminProperties = {
    guildId: string;
    height: number;
};

function SoundAdmin({ guildId, height }: SoundAdminProperties) {
    const [isLoading, setIsLoading] = useState(false);
    const {
        isLoading: isSoundListLoading,
        categories,
        sounds,
        updateCategory,
        updateSound,
        removeCategory,
        removeSound,
    } = UseSoundList(guildId);
    const [activeCategory, setActiveCategory] = useState<Category | null>(null);

    useEffect(() => {
        if (!activeCategory) {
            setActiveCategory(categories[0]);
        }
    }, [categories]);

    const addCategory = async () => {
        if (isLoading || isSoundListLoading) return;
        const name = prompt('Enter category name:')?.trim();
        if (!name) return;
        setIsLoading(true);
        updateCategory(
            await api.saveCategory({
                name,
                guildId,
                sort: (categories.length + 1) * 1000,
            })
        );
        setIsLoading(false);
    };

    const deleteCategory = async (category: Category) => {
        if (isLoading || isSoundListLoading) return;
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
        removeCategory(category);
        setIsLoading(false);
    };

    const renameCategory = async (category: Category) => {
        if (isLoading || isSoundListLoading) return;
        const name = prompt('Enter category name:', category.name)?.trim();
        if (!name) return;
        setIsLoading(true);
        category.name = name;
        updateCategory(await api.saveCategory(category));
        setIsLoading(false);
    };

    const onSortCategory = async (fromIndex: number, toIndex: number) => {
        let updatedCategories = Array.from(categories);
        const item = updatedCategories.splice(fromIndex, 1);
        updatedCategories = [
            ...updatedCategories.slice(0, toIndex),
            ...item,
            ...updatedCategories.slice(toIndex),
        ];
        updatedCategories.forEach((category, index) => {
            category.sort = index;
        });
        await api.sortCategories(updatedCategories);
        updateCategory(updatedCategories[0]);
    };

    const addSound = async () => {
        if (isLoading || isSoundListLoading) return;
        document.getElementById('sound-admin-file').click();
    };

    const onFile = async () => {
        if (!activeCategory || isLoading || isSoundListLoading) return;

        setIsLoading(true);
        const fileElement = document.getElementById(
            'sound-admin-file'
        ) as HTMLInputElement;
        if (fileElement.files) {
            // TODO try/catch doesn't catch errors??
            let soundData = null;
            try {
                soundData = await soundUtils.convert(
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
                sort: (sounds.length + 1) * 1000,
            });
            updateSound(sound);
        }
        setIsLoading(false);
    };

    const renameSound = async (sound: Sound) => {
        if (isLoading || isSoundListLoading) return;
        const name = prompt('Enter sound name:', sound.name)?.trim();
        if (!name) return;
        setIsLoading(true);
        sound.name = name;
        updateSound(await api.saveSound(sound));
        setIsLoading(false);
    };

    const deleteSound = async (sound: Sound) => {
        if (isLoading || isSoundListLoading) return;
        setIsLoading(true);
        await api.deleteSound(sound);
        removeSound(sound);
        setIsLoading(false);
    };

    const onSortSound = async (fromIndex: number, toIndex: number) => {
        const sortSounds = sounds.filter(
            (sound) => sound.categoryId === activeCategory.id
        );
        const item = sortSounds.splice(fromIndex, 1);
        const updatedSounds = [
            ...sortSounds.slice(0, toIndex),
            ...item,
            ...sortSounds.slice(toIndex),
        ];
        updatedSounds.forEach((sound, index) => {
            sound.sort = index;
        });
        await api.sortSounds(updatedSounds);
        updateSound(updatedSounds[0]);
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
            <div className="categories" style={{ height: `${height - 150}px` }}>
                <div className="section-title">Categories</div>
                {activeCategory && (
                    <SortableList
                        onSortEnd={onSortCategory}
                        className="category-sort"
                        draggedItemClassName="category-dragged"
                        lockAxis="y"
                    >
                        {categories.map((category) => (
                            <SortableItem key={`sort-category-${category.id}`}>
                                <div>
                                    <SoundAdminOption
                                        key={`category-${category.id}`}
                                        active={
                                            !isLoading &&
                                            activeCategory.id === category.id
                                        }
                                        label={category.name}
                                        onClick={() =>
                                            setActiveCategory(category)
                                        }
                                        onDelete={() =>
                                            deleteCategory(category)
                                        }
                                        onEdit={() => renameCategory(category)}
                                    />
                                </div>
                            </SortableItem>
                        ))}
                    </SortableList>
                )}
            </div>
            <div className="sounds" style={{ height: `${height - 150}px` }}>
                <div className="section-title">Sounds</div>
                {activeCategory && (
                    <SortableList
                        onSortEnd={onSortSound}
                        className="sound-sort"
                    >
                        {sounds
                            .filter(
                                (sound) =>
                                    sound.categoryId === activeCategory.id
                            )
                            .map((sound) => (
                                <SortableItem key={`sort-sound-${sound.id}`}>
                                    <div>
                                        <SoundAdminOption
                                            key={`sound-${sound.id}`}
                                            label={sound.name}
                                            onEdit={() => renameSound(sound)}
                                            onDelete={() => deleteSound(sound)}
                                        />
                                    </div>
                                </SortableItem>
                            ))}
                    </SortableList>
                )}
            </div>
        </div>
    );
}

export default SoundAdmin;
