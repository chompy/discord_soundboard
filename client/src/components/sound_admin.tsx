import React, { useState, useEffect } from 'react';

import { api, Category, Sound } from '../api';
import Button from './button';
import SoundAdminOption from './sound_admin_option';
import { sound as soundUtils } from '../sound';
import { error } from '../utils';
import SortableList, { SortableItem } from 'react-easy-sort';
import { SoundList } from '../hooks/sound_list';

export type SoundAdminProperties = {
    guildId: string;
    height: number;
    soundList: SoundList;
};

function SoundAdmin({ guildId, height, soundList }: SoundAdminProperties) {
    const [isLoading, setIsLoading] = useState(false);
    const {
        isLoading: isSoundListLoading,
        categories,
        sounds,
        localRefresh,
    } = soundList;
    const [activeCategory, setActiveCategory] = useState<Category | null>(null);

    useEffect(() => {
        if (!activeCategory) {
            setActiveCategory(categories[0]);
        }
    }, [categories]);

    const setIsLoadingTimeout = (value?: boolean) =>
        setTimeout(() => setIsLoading(value), 250);

    const addCategory = async () => {
        if (isLoading || isSoundListLoading) return;
        const name = prompt('Enter category name:')?.trim();
        if (!name) return;
        setIsLoading(true);
        categories.push(
            await api.saveCategory({
                name,
                guildId,
                sort: (categories.length + 1) * 1000,
            })
        );
        localRefresh();
        setIsLoadingTimeout();
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

        const index = categories.findIndex(
            (iterCategory) => iterCategory.id === category.id
        );
        categories.splice(index, 1);

        await api.deleteCategory(category);

        localRefresh();
        setIsLoadingTimeout();
    };

    const renameCategory = async (category: Category) => {
        if (isLoading || isSoundListLoading) return;
        const name = prompt('Enter category name:', category.name)?.trim();
        if (!name) return;
        setIsLoading(true);
        category.name = name;
        await api.saveCategory(category);
        localRefresh();
        setIsLoadingTimeout();
    };

    const onSortCategory = async (fromIndex: number, toIndex: number) => {
        if (isLoading || isSoundListLoading) return;

        setIsLoading(true);
        const category = categories[fromIndex];
        const moveTo = categories[toIndex];

        categories.forEach((category, index) => {
            category.sort = index + 1 * 1000;
        });
        category.sort = moveTo.sort + (fromIndex < toIndex ? 1 : -1);
        categories.sort((a, b) => a.sort - b.sort);

        await api.sortCategories(categories);
        localRefresh();
        setIsLoadingTimeout();
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
                setIsLoadingTimeout();
                return;
            }

            const hash = await api.uploadSound(soundData);
            const sound = await api.saveSound({
                name: fileElement.files[0].name,
                hash,
                categoryId: activeCategory.id,
                sort: (sounds.length + 1) * 1000,
            });
            sounds.push(sound);
            localRefresh();
        }
        setIsLoadingTimeout();
    };

    const renameSound = async (sound: Sound) => {
        if (isLoading || isSoundListLoading) return;
        const name = prompt('Enter sound name:', sound.name)?.trim();
        if (!name) return;
        setIsLoading(true);
        sound.name = name;
        await api.saveSound(sound);
        localRefresh();
        setIsLoadingTimeout();
    };

    const deleteSound = async (sound: Sound) => {
        if (isLoading || isSoundListLoading) return;
        setIsLoading(true);

        const index = sounds.findIndex(
            (iterSound) => iterSound.id === sound.id
        );
        sounds.splice(index, 1);

        await api.deleteSound(sound);

        localRefresh();
        setIsLoadingTimeout();
    };

    const onSortSound = async (fromIndex: number, toIndex: number) => {
        if (isLoading || isSoundListLoading) return;

        setIsLoading(true);

        sounds.forEach((sound, index) => {
            sound.sort = index + 1 * 1000;
        });

        const categorySounds = sounds.filter(
            (sound) => sound.categoryId === activeCategory.id
        );
        const moveFrom = categorySounds[fromIndex];
        const moveAfter = categorySounds[toIndex];

        moveFrom.sort = moveAfter.sort + (fromIndex < toIndex ? 1 : -1);

        sounds.sort((a, b) => a.sort - b.sort);

        await api.sortSounds(
            sounds.filter((sound) => sound.categoryId === activeCategory.id)
        );
        localRefresh();
        setIsLoadingTimeout();
    };

    return (
        <div className={`sound-admin${isLoading ? ' loading' : ''}`}>
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
