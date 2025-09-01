/** LOGGING **/
export const log = (message: string) => console.log(`> ${message}`);
export const error = (message: string) => console.error(`> ERROR: ${message}`);

/** MODAL **/
const modalElement = document.getElementById('modal');

export const modal = {
    open: (id: string) => {
        if (modalElement) modalElement.className = 'open';

        const modalContentElements =
            modalElement.getElementsByClassName('modal');
        if (modalContentElements) {
            for (let i = 0; i < modalContentElements.length; i++) {
                modalContentElements[i].classList.remove('open');
            }
        }

        const modalContentElement = document.getElementById(id);
        if (modalContentElement) {
            modalContentElement.classList.add('open');
        }
    },
    close: () => {
        if (modalElement) modalElement.className = '';
    },
};

if (modalElement)
    modalElement.addEventListener('click', (e) => {
        e.preventDefault();
        if (e.target && 'id' in e.target && e.target.id == 'modal') {
            modal.close();
        }
    });
