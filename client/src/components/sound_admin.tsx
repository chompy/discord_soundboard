import React, { useState, useEffect } from 'react';

import { api, Category, Sound } from '../api';
import Button from './button';
import SoundAdminOption from './sound_admin_option';
import { sound as soundUtils } from '../sound';
import { error } from '../utils';
import SortableList, { SortableItem, SortableKnob } from 'react-easy-sort';
import { SoundList } from '../hooks/sound_list';

export type SoundAdminProperties = {
    guildId: string;
    height: number;
    soundList: SoundList;
};

function SoundAdmin({ guildId, height, soundList }: SoundAdminProperties) {
    const [isLoading, setIsLoading] = useState(false);
    const { isLoading: isSoundListLoading, categories, sounds } = soundList;
    const [activeCategory, setActiveCategory] = useState<Category | null>(null);

    useEffect(() => {
        if (!activeCategory) {
            setActiveCategory(categories.get()[0]);
        }
    }, [categories]);

    const setIsLoadingTimeout = (value?: boolean) =>
        setTimeout(() => setIsLoading(value), 250);

    const addCategory = async () => {
        if (isLoading || isSoundListLoading) return;
        const name = prompt('Enter category name:')?.trim();
        if (!name) return;
        setIsLoading(true);
        categories.update(
            await api.saveCategory({
                name,
                guildId,
                sort: (categories.length() + 1) * 1000,
            })
        );
        setIsLoadingTimeout();
    };

    const deleteCategory = async (category: Category) => {
        if (isLoading || isSoundListLoading) return;

        const catSounds = sounds
            .get()
            .filter((sound) => sound.categoryId === category.id);

        if (
            catSounds.length > 0 &&
            !confirm(
                `Are you sure you want to delete this category? All ${catSounds.length} sound(s) will be deleted as well.`
            )
        ) {
            return;
        }

        setIsLoading(true);

        categories.delete(category);
        await api.deleteCategory(category);

        setIsLoadingTimeout();
    };

    const renameCategory = async (category: Category) => {
        if (isLoading || isSoundListLoading) return;
        const name = prompt('Enter category name:', category.name)?.trim();
        if (!name) return;
        setIsLoading(true);

        category.name = name;
        categories.update(category);
        await api.saveCategory(category);

        setIsLoadingTimeout();
    };

    const onSortCategory = async (fromIndex: number, toIndex: number) => {
        if (isLoading || isSoundListLoading) return;
        setIsLoading(true);
        const newCategories = categories.moveIndex(fromIndex, toIndex);
        await api.sortCategories(newCategories);
        setIsLoadingTimeout();
    };

    const listCategorySounds = (category: Category) =>
        sounds.get().filter((sound) => sound.categoryId === activeCategory.id);

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
                sort: (sounds.length() + 1) * 1000,
            });
            sounds.update(sound);
        }
        setIsLoadingTimeout();
    };

    const renameSound = async (sound: Sound) => {
        if (isLoading || isSoundListLoading) return;
        const name = prompt('Enter sound name:', sound.name)?.trim();
        if (!name) return;
        setIsLoading(true);

        sound.name = name;
        sounds.update(sound);
        await api.saveSound(sound);

        setIsLoadingTimeout();
    };

    const deleteSound = async (sound: Sound) => {
        if (isLoading || isSoundListLoading) return;
        setIsLoading(true);

        sounds.delete(sound);
        await api.deleteSound(sound);

        setIsLoadingTimeout();
    };

    const downloadSound = async (sound: Sound) => {        
        const data = await api.downloadSound(sound)
        const blob = new Blob([data], {type: "application/octet-stream"})
        const audioURL = window.URL.createObjectURL(blob);
        const link = document.createElement("a")
        link.href = audioURL;
        link.download = `${sound.name}.dat`
        link.click();
    }

    const onSortSound = async (fromIndex: number, toIndex: number) => {
        if (isLoading || isSoundListLoading) return;

        setIsLoading(true);

        const categorySounds = listCategorySounds(activeCategory);
        const fromSound = categorySounds[fromIndex];
        const toSound = categorySounds[toIndex];

        await api.sortSounds(
            sounds
                .move(fromSound, toSound)
                .filter((sound) => sound.categoryId === activeCategory.id)
        );
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
                        {categories.get().map((category) => (
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
                        {listCategorySounds(activeCategory).map((sound) => (
                            <SortableItem key={`sort-sound-${sound.id}`}>
                                <div>
                                    <SoundAdminOption
                                        key={`sound-${sound.id}`}
                                        label={sound.name}
                                        onEdit={() => renameSound(sound)}
                                        onDelete={() => deleteSound(sound)}
                                        onDownload={() => downloadSound(sound)}
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
