import { api, Category, Sound } from './api';
import { convertSound } from './converter';
import { log, error, modal } from './utils';

function populateSelect(
    element: HTMLSelectElement,
    options: { [key: string]: string }
) {
    element.innerHTML = '';
    for (const k in options) {
        const optionElement = document.createElement('option');
        optionElement.setAttribute('value', k);
        optionElement.innerText = options[k];
        element.append(optionElement);
    }
}

/** WEB ADMIN */
window.addEventListener('load', () => {
    let currentGuildId = '';
    let currentSoundHash = '';
    const currentCategories: Category[] = [];
    const currentSounds: Sound[] = [];

    const syncGuildData = async () => {
        currentCategories.splice(0);
        currentSounds.splice(0);
        if (!currentGuildId) return;
        currentCategories.push(...(await api.listCategories(currentGuildId)));
        currentSounds.push(...(await api.listSounds(currentGuildId)));
    };

    // -- ADD SOUND
    const modSoundNameInput = document.getElementById(
        'mod-sound-name'
    ) as HTMLInputElement;
    const modSoundCategoryInput = document.getElementById(
        'mod-sound-category'
    ) as HTMLSelectElement;
    const modSoundBtnSave = document.getElementById('mod-sound-btn-save');

    const addSoundBtn = document.getElementById('btn-add-sound');
    const uploadSoundInput = document.getElementById(
        'upload-sound'
    ) as HTMLInputElement;

    addSoundBtn.addEventListener('click', (e) => {
        e.preventDefault();
        uploadSoundInput.click();
    });

    uploadSoundInput.addEventListener('change', async (e) => {
        e.preventDefault();
        if (uploadSoundInput.files) {
            modal.open('modal-mod-sound');
            const opusFrames = await convertSound(
                await uploadSoundInput.files[0].arrayBuffer()
            );
            currentSoundHash = await api.uploadSound(opusFrames);
        }
    });

    modSoundBtnSave.addEventListener('click', async (e) => {
        e.preventDefault();
        if (!currentSoundHash) {
            alert('ERROR: Cannot locate sound upload.');
            modal.close();
            return;
        }
        const soundCategoryId = parseInt(modSoundCategoryInput.value);
        if (!soundCategoryId) {
            alert('ERROR: "Category" is required.');
            return;
        }
        const soundName = modSoundNameInput.value.trim();
        if (!soundName) {
            alert('ERROR: "Name" is required.');
            return;
        }
        await api.saveSound({
            name: soundName,
            hash: currentSoundHash,
            categoryId: soundCategoryId,
        });

        modal.close();
    });

    // -- SOUND LIST
    const soundListElement = document.getElementById('admin-sound-list');

    const soundListReset = () => {
        if (soundListElement) soundListElement.innerHTML = '';
    };
    const soundListCategory = (category: Category) => {
        if (!soundListElement) return;

        let categoryElement = document.getElementById(
            `category-${category.id}`
        );
        if (categoryElement) {
            return categoryElement;
        }

        categoryElement = document.createElement('div');
        categoryElement.className = 'category';
        categoryElement.id = `category-${category.id}`;
        const header = document.createElement('h3');
        header.innerText = category.name;
        categoryElement.append(header);

        const categorySoundsElement = document.createElement('ul');
        categorySoundsElement.id = `category-${category.id}-sounds`;
        categorySoundsElement.className = 'sound-list';
        categoryElement.append(categorySoundsElement);

        soundListElement.append(categoryElement);
        return categoryElement;
    };
    const soundListAddSound = (sound: Sound) => {
        if (!soundListElement) return;

        const category = currentCategories.find(
            (category) => category.id === sound.categoryId
        );
        if (!category) {
            error(`Could not find category for sound ${sound.id}`);
            return;
        }

        const categorySoundListElement =
            soundListCategory(category)?.getElementsByClassName(
                'sound-list'
            )[0];
        if (!categorySoundListElement) return;

        const soundElement = document.createElement('li');
        soundElement.innerText = sound.name;
        soundElement.id = `sound-${sound.id}`;

        categorySoundListElement.append(soundElement);
        return soundElement;
    };

    // -- GUILD SELECTION
    const changeCurrentGuild = async (guildId: string) => {
        log(`Change guild to ${guildId}`);
        currentGuildId = guildId;
        await syncGuildData();

        soundListReset();
        currentSounds.forEach((sound) => soundListAddSound(sound));

        populateSelect(
            modSoundCategoryInput,
            Object.fromEntries(
                currentCategories.map((catergory) => [
                    catergory.id,
                    catergory.name,
                ])
            )
        );
    };
    const guildSelectElement = document.getElementById(
        'guild-selection'
    ) as HTMLSelectElement;
    api.listGuilds().then((guilds) => {
        if (guilds && !currentGuildId) {
            changeCurrentGuild(guilds[0].id);
        }
        populateSelect(
            guildSelectElement,
            Object.fromEntries(guilds.map((guild) => [guild.id, guild.name]))
        );
    });
    guildSelectElement.addEventListener('change', (e) => {
        e.preventDefault();
        const guildId = guildSelectElement.getAttribute('value');
        if (!guildId) {
            error('unable to change guild, id not found');
            return;
        }
        changeCurrentGuild(guildId);
    });
});
